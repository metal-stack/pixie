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

package pixiecore

import (
	"errors"
	"fmt"
	"net"

	"go.universe.tf/netboot/dhcp4"
)

func (s *Server) serveDHCP(conn *dhcp4.Conn) {
	for {
		pkt, intf, err := conn.RecvDHCP()
		if err != nil {
			fmt.Println(err)
			return
		}
		if intf == nil {
			fmt.Println("no interface")
			continue
		}

		mach, isIpxe, fwtype, err := s.validateDHCP(pkt)
		if err != nil {
			fmt.Println(err)
			continue
		}

		spec, err := s.Booter.BootSpec(mach)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if spec == nil {
			fmt.Println("no spec")
			continue
		}

		// Machine should be booted.
		serverIP, err := interfaceIP(intf)
		if err != nil {
			fmt.Println(err)
			continue
		}

		resp, err := s.offerDHCP(pkt, mach, serverIP, isIpxe, fwtype)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if err = conn.SendDHCP(resp, intf); err != nil {
			fmt.Println(err)
			continue
		}
	}
}

func (s *Server) validateDHCP(pkt *dhcp4.Packet) (mach Machine, isIpxe bool, fwtype Firmware, err error) {
	if pkt.Type != dhcp4.MsgDiscover {
		return mach, false, 0, errors.New("not DHCPDISCOVER")
	}

	fwt, err := pkt.Options.Uint16(93)
	if err != nil {
		return mach, false, 0, fmt.Errorf("malformed arch: %s", err)
	}
	fwtype = Firmware(fwt)
	if s.Ipxe[fwtype] == nil {
		return mach, false, 0, fmt.Errorf("unsupported firmware type %d", fwtype)
	}

	guid := pkt.Options[97]
	switch len(guid) {
	case 0:
		// A missing GUID is invalid according to the spec, however
		// there are PXE ROMs in the wild that omit the GUID and still
		// expect to boot.
	case 17:
		if guid[0] != 0 {
			return mach, false, 0, errors.New("malformed GUID (leading byte must be zero)")
		}
	default:
		return mach, false, 0, errors.New("malformed GUID (wrong size)")
	}

	// iPXE options
	if len(pkt.Options[175]) > 0 {
		bs := pkt.Options[175]
		for len(bs) > 0 {
			if len(bs) < 2 || len(bs)-2 < int(bs[1]) {
				return mach, false, 0, errors.New("Malformed iPXE option")
			}
			switch bs[0] {
			case 19:
				// This iPXE build supports HTTP, so is appropriate
				// for going straight into the OS kernel, no need to
				// chainload our own.
				isIpxe = true
			}
			bs = bs[2+int(bs[1]):]
		}
	}

	mach.MAC = pkt.HardwareAddr
	mach.Arch = fwToArch[fwtype]
	return mach, isIpxe, fwtype, nil
}

func (s *Server) offerDHCP(pkt *dhcp4.Packet, mach Machine, serverIP net.IP, isIpxe bool, fwtype Firmware) (*dhcp4.Packet, error) {
	resp := &dhcp4.Packet{
		Type:          dhcp4.MsgOffer,
		TransactionID: pkt.TransactionID,
		Broadcast:     true,
		HardwareAddr:  mach.MAC,
		RelayAddr:     pkt.RelayAddr,
		ServerAddr:    serverIP,
		Options:       make(dhcp4.Options),
	}
	resp.Options[dhcp4.OptServerIdentifier] = serverIP
	// says the server should identify itself as a PXEClient vendor
	// type, even though it's a server. Strange.
	resp.Options[dhcp4.OptVendorIdentifier] = []byte("PXEClient")
	if pkt.Options[97] != nil {
		resp.Options[97] = pkt.Options[97]
	}

	// TODO: for maximum support, need to also send a BINL response to
	// UEFI clients, or they might ignore this ProxyDHCP response.

	if isIpxe {
		resp.BootFilename = fmt.Sprintf("http://%s:%d/_/ipxe?arch=%d&mac=%s", serverIP, s.HTTPPort, mach.Arch, mach.MAC)
	} else {
		resp.BootServerName = serverIP.String()
		resp.BootFilename = fmt.Sprintf("%d", fwtype)
	}

	if fwtype == FirmwareX86PC {
		// In theory, the PXE boot options are required by PXE
		// clients. However, some UEFI firmwares don't actually
		// support PXE properly, and will ignore ProxyDHCP responses
		// that include the option.
		//
		// On the other hand, seemingly all firmwares support a
		// variant of the protocol where option 43 is not
		// provided. They behave as if option 43 had pointed them to a
		// PXE boot server on port 4011 of the machine sending the
		// ProxyDHCP response. Looking at TianoCore sources, I believe
		// this is the BINL protocol, which is Microsoft-specific and
		// lacks a specification. However, empirically, this code
		// seems to work.
		//
		// But this code block is for classic old BIOS, which does
		// behave and want option 43 to tell it how to boot.

		pxe := dhcp4.Options{
			// PXE Boot Server Discovery Control - bypass, just boot from filename.
			6: []byte{8},
		}
		bs, err := pxe.Marshal()
		if err != nil {
			return nil, fmt.Errorf("failed to serialize PXE vendor options: %s", err)
		}
		resp.Options[43] = bs
	}
	return resp, nil
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

	return nil, fmt.Errorf("interface %s has no usable server addresses", intf.Name)
}
