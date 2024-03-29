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

	"github.com/metal-stack/pixie/dhcp4"
	"golang.org/x/net/ipv4"
)

// TODO: this may actually be the BINL protocol, a
// Microsoft-proprietary fork of PXE that is more universally
// supported in UEFI than PXE itself. Need to comb through the
// TianoCore EDK2 source code to figure out if what this is doing is
// actually BINL, and if so rename everything.

func (s *Server) servePXE(conn net.PacketConn) error {
	buf := make([]byte, 1024)
	l := ipv4.NewPacketConn(conn)
	if err := l.SetControlMessage(ipv4.FlagInterface, true); err != nil {
		return fmt.Errorf("couldn't get interface metadata on PXE port: %w", err)
	}

	for {
		n, msg, addr, err := l.ReadFrom(buf)
		if err != nil {
			return fmt.Errorf("receiving packet: %w", err)
		}

		pkt, err := dhcp4.Unmarshal(buf[:n])
		if err != nil {
			s.Log.Debug("Packet is not a DHCP packet", "addr", addr, "error", err)
			continue
		}

		if err = s.isBootDHCP(pkt); err != nil {
			s.Log.Debug("Ignoring packet", "mac", pkt.HardwareAddr, "addr", addr, "error", err)
		}
		fwtype, err := s.validatePXE(pkt)
		if err != nil {
			s.Log.Info("Unusable packet", "mac", pkt.HardwareAddr, "addr", addr, "error", err)
			continue
		}

		intf, err := net.InterfaceByIndex(msg.IfIndex)
		if err != nil {
			s.Log.Info("Couldn't get information about local network interface", "ifindex", msg.IfIndex, "error", err)
			continue
		}

		serverIP, err := interfaceIP(intf)
		if err != nil {
			s.Log.Info("Want to boot, but couldn't get a source address", "mac", pkt.HardwareAddr, "addr", addr, "interface", intf.Name, "error", err)
			continue
		}

		s.machineEvent(pkt.HardwareAddr, machineStatePXE, "Sent PXE configuration")

		resp := s.offerPXE(pkt, serverIP, fwtype)

		bs, err := resp.Marshal()
		if err != nil {
			s.Log.Info("Failed to marshal PXE offer", "mac", pkt.HardwareAddr, "addr", addr, "error", err)
			continue
		}

		if _, err := l.WriteTo(bs, &ipv4.ControlMessage{
			IfIndex: msg.IfIndex,
		}, addr); err != nil {
			s.Log.Info("Failed to send PXE response", "mac", pkt.HardwareAddr, "to", addr, "error", err)
		}
	}
}

func (s *Server) validatePXE(pkt *dhcp4.Packet) (fwtype Firmware, err error) {
	fwt, err := pkt.Options.Uint16(93)
	if err != nil {
		return 0, fmt.Errorf("malformed DHCP option 93 (required for PXE): %w", err)
	}
	// see: https://ipxe.org/cfg/platform for reference
	switch fwt {
	case 0:
		fwtype = FirmwareX86Ipxe
	case 6:
		fwtype = FirmwareEFI32
	case 7:
		fwtype = FirmwareEFI64
	case 9:
		fwtype = FirmwareEFIBC
	default:
		return 0, fmt.Errorf("unsupported client firmware type '%d' (please file a bug!)", fwt)
	}
	if s.Ipxe[fwtype] == nil {
		return 0, fmt.Errorf("unsupported client firmware type '%d' (please file a bug!)", fwtype)
	}

	guid := pkt.Options[97]
	switch len(guid) {
	case 0:
		// Accept missing GUIDs even though it's a spec violation,
		// same as in dhcp.go.
	case 17:
		if guid[0] != 0 {
			return 0, errors.New("malformed client GUID (option 97), leading byte must be zero")
		}
	default:
		return 0, errors.New("malformed client GUID (option 97), wrong size")
	}

	return fwtype, nil
}

func (s *Server) offerPXE(pkt *dhcp4.Packet, serverIP net.IP, fwtype Firmware) (resp *dhcp4.Packet) {
	resp = &dhcp4.Packet{
		Type:           dhcp4.MsgAck,
		TransactionID:  pkt.TransactionID,
		HardwareAddr:   pkt.HardwareAddr,
		ClientAddr:     pkt.ClientAddr,
		RelayAddr:      pkt.RelayAddr,
		ServerAddr:     serverIP,
		BootServerName: serverIP.String(),
		BootFilename:   fmt.Sprintf("%s/%d", pkt.HardwareAddr, fwtype),
		Options: dhcp4.Options{
			dhcp4.OptServerIdentifier: serverIP,
			dhcp4.OptVendorIdentifier: []byte("PXEClient"),
		},
	}
	if pkt.Options[97] != nil {
		resp.Options[97] = pkt.Options[97]
	}

	return resp
}
