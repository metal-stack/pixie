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

//+build linux

package dhcp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"

	"golang.org/x/net/bpf"
	"golang.org/x/net/ipv4"
)

const (
	udpProtocolNumber = 17
)

type linuxConn struct {
	port uint16
	conn *ipv4.RawConn
}

func init() {
	platformConn = newLinuxConn
}

func newLinuxConn(addr string) (Conn, error) {
	if addr == "" {
		addr = ":67"
	}
	udpAddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return nil, err
	}
	if udpAddr.IP != nil && udpAddr.IP.To4() == nil {
		return nil, fmt.Errorf("%s is not an IPv4 address", addr)
	}
	udpAddr.IP = udpAddr.IP.To4()
	if udpAddr.Port == 0 {
		return nil, fmt.Errorf("%s must specify a listen port", addr)
	}

	filter, err := bpf.Assemble([]bpf.Instruction{
		// Load IPv4 packet length
		bpf.LoadMemShift{Off: 0},
		// Get UDP dport
		bpf.LoadIndirect{Off: 2, Size: 2},
		// Correct dport?
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: uint32(udpAddr.Port), SkipFalse: 1},
		// Accept
		bpf.RetConstant{Val: 1500},
		// Ignore
		bpf.RetConstant{Val: 0},
	})
	if err != nil {
		return nil, err
	}

	c, err := net.ListenPacket("ip4:17", udpAddr.IP.String())
	if err != nil {
		return nil, err
	}
	r, err := ipv4.NewRawConn(c)
	if err != nil {
		c.Close()
		return nil, err
	}
	if err = r.SetControlMessage(ipv4.FlagInterface, true); err != nil {
		c.Close()
		return nil, fmt.Errorf("setting packet filter: %s", err)
	}
	if err = r.SetBPF(filter); err != nil {
		c.Close()
		return nil, fmt.Errorf("setting packet filter: %s", err)
	}

	ret := &linuxConn{
		port: uint16(udpAddr.Port),
		conn: r,
	}
	return ret, nil
}

func (c *linuxConn) Close() error {
	return c.conn.Close()
}

func (c *linuxConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *linuxConn) RecvDHCP() (*Packet, *net.Interface, error) {
	var buf [1500]byte
	for {
		_, p, cm, err := c.conn.ReadFrom(buf[:])
		if err != nil {
			return nil, nil, err
		}
		if len(p) < 8 {
			continue
		}
		pkt, err := Unmarshal(p[8:])
		if err != nil {
			continue
		}
		intf, err := net.InterfaceByIndex(cm.IfIndex)
		if err != nil {
			return nil, nil, err
		}
		// TODO: possibly more validation that the IPv4 header lines
		// up with what the packet.
		return pkt, intf, nil
	}
}

func (c *linuxConn) SendDHCP(pkt *Packet, intf *net.Interface) error {
	b, err := pkt.Marshal()
	if err != nil {
		return err
	}

	raw := make([]byte, 8+len(b))
	// src port
	binary.BigEndian.PutUint16(raw[:2], c.port)
	// dst port
	binary.BigEndian.PutUint16(raw[2:4], uint16(dhcpClientPort))
	// length
	binary.BigEndian.PutUint16(raw[4:6], uint16(8+len(b)))
	copy(raw[8:], b)

	hdr := ipv4.Header{
		Version:  4,
		Len:      ipv4.HeaderLen,
		TOS:      0xc0, // DSCP CS6 (Network Control)
		TotalLen: ipv4.HeaderLen + 8 + len(b),
		TTL:      64,
		Protocol: 17,
	}

	switch pkt.TxType() {
	case TxBroadcast, TxHardwareAddr:
		hdr.Dst = net.IPv4bcast
		cm := ipv4.ControlMessage{
			IfIndex: intf.Index,
		}
		return c.conn.WriteTo(&hdr, raw, &cm)
	case TxRelayAddr:
		// Send to the server port, not the client port.
		binary.BigEndian.PutUint16(raw[2:4], 67)
		hdr.Dst = pkt.RelayAddr
		return c.conn.WriteTo(&hdr, raw, nil)
	case TxClientAddr:
		hdr.Dst = pkt.ClientAddr
		return c.conn.WriteTo(&hdr, raw, nil)
	default:
		return errors.New("unknown TX type for packet")
	}
}
