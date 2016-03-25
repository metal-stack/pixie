// Copyright 2016 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dhcp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"runtime"
	"unsafe"

	"golang.org/x/sys/unix"
)

var (
	protoAll = int(unix.ETH_P_ALL)
	macbcast = []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
)

func init() {
	// This kernel API is icky through and through. It wants the
	// ethernet protocol number in big-endian form, but receives it as
	// a native-endian integer. Thus, we need to do the moral
	// equivalent of htons().
	i := uint16(1)
	b := *(*byte)(unsafe.Pointer(&i))
	if b == 1 {
		protoAll = protoAll << 8
	}
}

// LinuxConn implements Conn using Linux raw sockets.
//
// The advantage compared to PortableConn is that LinuxConn does not
// need to bind to any port, and so can be run alongside other DHCP
// services on the same machine. However, using it requires
// CAP_NET_RAW, whereas PortableConn doesn't.
type LinuxConn struct {
	port       int
	ethernetFd int // AF_PACKET SOCK_RAW socket
	ipFd       int // AF_INET SOCK_RAW IPPROTO_RAW socket
}

func closeLinuxConn(c *LinuxConn) {
	if c.ethernetFd != -1 {
		unix.Close(c.ethernetFd)
		c.ethernetFd = -1
	}
	if c.ipFd != -1 {
		unix.Close(c.ipFd)
		c.ipFd = -1
	}
}

// NewLinuxConn creates a LinuxConn that receives DHCP packets. TODO better
func NewLinuxConn(addr string) (*LinuxConn, error) {
	if addr == "" {
		addr = ":67"
	}
	udpAddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return nil, err
	}
	udpAddr.IP = udpAddr.IP.To4()

	eth, err := unix.Socket(unix.AF_PACKET, unix.SOCK_RAW, protoAll)
	if err != nil {
		return nil, err
	}

	filter := filterPortOnly(udpAddr.Port)
	if udpAddr.IP != nil {
		filter = filterIPAndPort(udpAddr.IP, udpAddr.Port)
	}
	if err = attachFilter(eth, filter); err != nil {
		unix.Close(eth)
		return nil, err
	}

	ip, err := unix.Socket(unix.AF_INET, unix.SOCK_RAW, unix.IPPROTO_RAW)
	if err != nil {
		unix.Close(eth)
		return nil, err
	}

	// if err = attachFilter(ip, filterNoPackets()); err != nil {
	// 	unix.Close(eth)
	// 	unix.Close(ip)
	// 	return nil, err
	// }

	ret := &LinuxConn{
		port:       udpAddr.Port,
		ethernetFd: eth,
		ipFd:       ip,
	}
	runtime.SetFinalizer(ret, closeLinuxConn)
	return ret, nil
}

// Close closes the connection.
func (c *LinuxConn) Close() error {
	closeLinuxConn(c)
	return nil
}

// RecvDHCP implements the Conn RecvDHCP method.
func (c *LinuxConn) RecvDHCP() (*Packet, *net.Interface, error) {
	buf := make([]byte, 1500)
	for {
		n, from, err := unix.Recvfrom(c.ethernetFd, buf, 0)
		if err != nil {
			return nil, nil, err
		}
		bs := buf[:n]
		// Advance past the ethernet, IP and UDP headers, to reach the
		// DHCP packet.
		off := 22 + 4*int(buf[14]&0xf)
		pkt, err := Unmarshal(bs[off:])
		if err != nil {
			// TODO: return temporary error to allow the server to log
			// stuff.
			continue
		}

		if err = validatePacket(bs, pkt); err != nil {
			// TODO: return temporary error to allow the server to log
			// stuff.
			continue
		}

		addr := from.(*unix.SockaddrLinklayer)
		intf, err := net.InterfaceByIndex(addr.Ifindex)
		if err != nil {
			return nil, nil, err
		}

		return pkt, intf, nil
	}
}

// SendDHCP implements the Conn SendDHCP method.
func (c *LinuxConn) SendDHCP(pkt *Packet, intf *net.Interface) error {
	payload, err := pkt.Marshal()
	if err != nil {
		return err
	}
	switch pkt.TxType() {
	case TxBroadcast:
		if intf == nil {
			return errors.New("packet needs to be broadcast, but no interface specified")
		}
		srcIP, err := interfaceIP(intf)
		if err != nil {
			return err
		}
		bs := assemblePacket(intf.HardwareAddr, macbcast, srcIP, net.IPv4bcast, c.port, 68, payload)
		addr := &unix.SockaddrLinklayer{
			Ifindex: intf.Index,
			Halen:   6,
		}
		copy(addr.Addr[:6], intf.HardwareAddr)
		if err = unix.Sendto(c.ethernetFd, bs, 0, addr); err != nil {
			return err
		}
	case TxRelayAddr:
		bs := assemblePacket(nil, nil, nil, pkt.RelayAddr, c.port, 67, payload)
		bs = bs[14:] // Skip the ethernet header
		addr := &unix.SockaddrInet4{}
		copy(addr.Addr[:], pkt.RelayAddr.To4())
		if err = unix.Sendto(c.ipFd, bs, 0, addr); err != nil {
			return err
		}
	case TxClientAddr:
		bs := assemblePacket(nil, nil, nil, pkt.ClientAddr, c.port, 68, payload)
		bs = bs[14:] // Skip the ethernet header
		addr := &unix.SockaddrInet4{}
		if err = unix.Sendto(c.ipFd, bs, 0, addr); err != nil {
			return err
		}
	case TxHardwareAddr:
		if intf == nil {
			return errors.New("packet needs to be transmitted to unconfigured client, but no interface specified")
		}
		srcIP, err := interfaceIP(intf)
		if err != nil {
			return err
		}
		bs := assemblePacket(intf.HardwareAddr, pkt.HardwareAddr, srcIP, pkt.YourAddr, c.port, 68, payload)

		addr := &unix.SockaddrLinklayer{
			Ifindex: intf.Index,
			Halen:   6,
		}
		copy(addr.Addr[:6], intf.HardwareAddr)
		if err = unix.Sendto(c.ethernetFd, bs, 0, addr); err != nil {
			return err
		}
	}

	return nil
}

func validatePacket(frame []byte, pkt *Packet) error {
	if pkt.RelayAddr != nil {
		// If the packet is from a relay, no validation is needed.
		return nil
	}

	// ciaddr must match the IP header's source IP (either an actual
	// address, or 0.0.0.0).
	if !bytes.Equal(pkt.ClientAddr, frame[24:28]) {
		return errors.New("ciaddr doesn't match packet source IP")
	}
	// chaddr must match the source MAC address
	if !bytes.Equal(pkt.HardwareAddr, frame[6:12]) {
		return errors.New("chaddr doesn't match packet source MAC")
	}

	return nil
}

func interfaceIP(intf *net.Interface) (net.IP, error) {
	addrs, err := intf.Addrs()
	if err != nil {
		return nil, err
	}

	// Try to find an IPv4 address to use, in the following order:
	// global unicast (includes rfc1918), link-local unicast,
	// loopback.
	fs := [](func(net.IP) bool){
		net.IP.IsGlobalUnicast,
		net.IP.IsLinkLocalUnicast,
		net.IP.IsLoopback,
	}
	for _, f := range fs {
		for _, a := range addrs {
			ipaddr, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipaddr.IP.To4()
			if ip == nil {
				continue
			}
			if f(ip) {
				return ip, nil
			}
		}
	}

	return nil, fmt.Errorf("interface %s has no unicast address usable as a DHCP packet source", intf.Name)
}

func assemblePacket(srcMAC, dstMAC net.HardwareAddr, srcIP, dstIP net.IP, srcPort, dstPort int, payload []byte) []byte {
	buf := make([]byte, 42, 42+len(payload))

	// Ethernet header
	copy(buf[:6], dstMAC)
	copy(buf[6:12], srcMAC)
	binary.BigEndian.PutUint16(buf[12:14], 0x0800)

	// IP header
	buf[14] = (4 << 4) + 5                                          // IP version 4, 5-word header (20b)
	buf[15] = 0xc0                                                  // ToS CS6 (Network Control)
	binary.BigEndian.PutUint16(buf[16:18], uint16(28+len(payload))) // IP packet length
	binary.BigEndian.PutUint32(buf[18:22], 0x4000)                  // ID=0, frag_off=0, dont_fragment=1
	buf[22] = 64                                                    // TTL
	buf[23] = 17                                                    // Inner protocol: UDP
	copy(buf[26:30], srcIP.To4())
	copy(buf[30:34], dstIP.To4())
	fmt.Println(buf[30:34])

	var cksum uint32
	for i := 14; i < 34; i += 2 {
		cksum += uint32(binary.BigEndian.Uint16(buf[i : i+2]))
	}
	cksum = (cksum >> 16) + (cksum & 0xFFFF)
	binary.BigEndian.PutUint16(buf[24:26], ^uint16(cksum))

	// UDP header
	binary.BigEndian.PutUint16(buf[34:36], uint16(srcPort))        // Source port
	binary.BigEndian.PutUint16(buf[36:38], uint16(dstPort))        // Destination port
	binary.BigEndian.PutUint16(buf[38:40], uint16(8+len(payload))) // UDP length

	return append(buf, payload...)
}

func attachFilter(fd int, filter *unix.SockFprog) error {
	_, _, errno := unix.Syscall6(unix.SYS_SETSOCKOPT, uintptr(fd), uintptr(unix.SOL_SOCKET), uintptr(unix.SO_ATTACH_FILTER), uintptr(unsafe.Pointer(filter)), uintptr(unsafe.Sizeof(*filter)), 0)
	if errno != 0 {
		return errno
	}
	return nil
}

func filterPortOnly(port int) *unix.SockFprog {
	// This filter comes from:
	// tcpdump -dd 'ip and udp dst port 68'
	filter := []unix.SockFilter{
		{0x28, 0, 0, 12},           // Load ethernet frame type
		{0x15, 0, 8, 0x0800},       // Is IPv4?
		{0x30, 0, 0, 23},           // Load IP packet type
		{0x15, 0, 6, 17},           // Is UDP?
		{0x28, 0, 0, 20},           // Load fragment offset
		{0x45, 4, 0, 0x1fff},       // Is first/only fragment?
		{0xb1, 0, 0, 14},           // Jump to start of UDP header
		{0x48, 0, 0, 16},           // Load destination port
		{0x15, 0, 1, uint32(port)}, // Is correct port?
		{0x6, 0, 0, 0x40000},       // Yes, receive packet
		{0x6, 0, 0, 0},             // No, ignore packet
	}
	return &unix.SockFprog{
		Len:    uint16(len(filter)),
		Filter: &filter[0],
	}
}

func filterIPAndPort(dstIP net.IP, port int) *unix.SockFprog {
	d := binary.BigEndian.Uint32([]byte(dstIP.To4()))
	// This filter comes from:
	// tcpdump -dd 'ip and udp and (dst 192.168.2.2 or dst 255.255.255.255) and dst port 68'
	filter := []unix.SockFilter{
		{0x28, 0, 0, 12},           // Load ethernet frame type
		{0x15, 0, 11, 0x0800},      // Is IPv4?
		{0x30, 0, 0, 23},           // Load IP packet type
		{0x15, 0, 9, 17},           // Is UDP?
		{0x20, 0, 0, 30},           // Load destination IP
		{0x15, 1, 0, d},            // Is target IP?
		{0x15, 0, 6, 0xffffffff},   // Is Broadcast?
		{0x28, 0, 0, 20},           // Load fragment offset
		{0x45, 4, 0, 0x1fff},       // Is first/only fragment?
		{0xb1, 0, 0, 14},           // Jump to start of UDP header
		{0x48, 0, 0, 16},           // Load destination port
		{0x15, 0, 1, uint32(port)}, // Is correct port?
		{0x6, 0, 0, 0x40000},       // Yes, receive packet
		{0x6, 0, 0, 0},             // No, ignore packet
	}
	return &unix.SockFprog{
		Len:    uint16(len(filter)),
		Filter: &filter[0],
	}
}

func filterNoPackets() *unix.SockFprog {
	filter := []unix.SockFilter{
		{0x6, 0, 0, 0}, // ignore packet
	}
	return &unix.SockFprog{
		Len:    uint16(len(filter)),
		Filter: &filter[0],
	}
}
