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
	"golang.org/x/net/ipv4"
)

func (s *Server) servePXE(conn net.PacketConn) {
	buf := make([]byte, 1024)
	l := ipv4.NewPacketConn(conn)
	if err := l.SetControlMessage(ipv4.FlagInterface, true); err != nil {
		fmt.Println(err)
		return
	}

	for {
		n, msg, addr, err := l.ReadFrom(buf)
		if err != nil {
			fmt.Println(err)
			return
		}

		pkt, err := dhcp4.Unmarshal(buf[:n])
		if err != nil {
			fmt.Println("not a DHCP packet")
			continue
		}

		fwtype, err := s.validatePXE(pkt)
		if err != nil {
			fmt.Println(err)
			continue
		}

		intf, err := net.InterfaceByIndex(msg.IfIndex)
		if err != nil {
			fmt.Println(err)
			continue
		}

		serverIP, err := interfaceIP(intf)
		if err != nil {
			fmt.Printf("Couldn't find an IP address to use to reply to %s: %s\n", addr, err)
			continue
		}

		resp, err := s.offerPXE(pkt, serverIP, fwtype)
		if err != nil {
			fmt.Println(err)
			continue
		}

		bs, err := resp.Marshal()
		if err != nil {
			fmt.Println(err)
		}

		if _, err := l.WriteTo(bs, &ipv4.ControlMessage{
			IfIndex: msg.IfIndex,
		}, addr); err != nil {
			fmt.Println(err)
		}
	}
}

func (s *Server) validatePXE(pkt *dhcp4.Packet) (fwtype Firmware, err error) {
	if pkt.Type != dhcp4.MsgRequest {
		return 0, errors.New("not DHCPREQUEST")
	}

	fwt, err := pkt.Options.Uint16(93)
	if err != nil {
		return 0, fmt.Errorf("malformed arch: %s", err)
	}
	fwtype = Firmware(fwt)
	if s.Ipxe[fwtype] == nil {
		return 0, fmt.Errorf("unsupported firmware type %d", fwtype)
	}

	guid := pkt.Options[97]
	switch len(guid) {
	case 0:
		// A missing GUID is invalid according to the spec, however
		// there are PXE ROMs in the wild that omit the GUID and still
		// expect to boot.
	case 17:
		if guid[0] != 0 {
			return 0, errors.New("malformed GUID (leading byte must be zero)")
		}
	default:
		return 0, errors.New("malformed GUID (wrong size)")
	}

	return fwtype, nil
}

func (s *Server) offerPXE(pkt *dhcp4.Packet, serverIP net.IP, fwtype Firmware) (resp *dhcp4.Packet, err error) {
	resp = &dhcp4.Packet{
		Type:          dhcp4.MsgAck,
		TransactionID: pkt.TransactionID,
		//Broadcast:      true,
		HardwareAddr:   pkt.HardwareAddr,
		ClientAddr:     pkt.ClientAddr,
		RelayAddr:      pkt.RelayAddr,
		ServerAddr:     serverIP,
		BootServerName: serverIP.String(),
		BootFilename:   fmt.Sprintf("%d", fwtype),
		Options: dhcp4.Options{
			dhcp4.OptServerIdentifier: serverIP,
			dhcp4.OptVendorIdentifier: []byte("PXEClient"),
			//dhcp4.OptTFTPServer:       []byte(serverIP.String()),
			//dhcp4.OptBootFile:         []byte(fmt.Sprintf("%d", fwtype)),
		},
	}
	if pkt.Options[97] != nil {
		resp.Options[97] = pkt.Options[97]
	}
	// pxe := dhcp4.Options{
	// 	// PXE Boot Server Discovery Control - bypass, just boot from filename.
	// 	6: []byte{8},
	// }
	// bs, err := pxe.Marshal()
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to serialize PXE vendor options: %s", err)
	// }
	// resp.Options[43] = bs

	return resp, nil
}
