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
	"errors"
	"net"
	"time"

	"golang.org/x/net/ipv4"
)

// defined as a var so tests can override it.
var dhcpClientPort = 68

var platformConn func(string) (Conn, error)

func NewConn(addr string) (Conn, error) {
	if platformConn != nil {
		c, err := platformConn(addr)
		if err == nil {
			return c, nil
		}
	}
	// Always try falling back to the portable implementation
	return newPortableConn(addr)
}

type portableConn struct {
	conn *ipv4.PacketConn
}

func newPortableConn(addr string) (Conn, error) {
	c, err := net.ListenPacket("udp4", addr)
	if err != nil {
		return nil, err
	}
	l := ipv4.NewPacketConn(c)
	if err = l.SetControlMessage(ipv4.FlagInterface, true); err != nil {
		l.Close()
		return nil, err
	}
	return &portableConn{l}, nil
}

func (c *portableConn) Close() error {
	return c.conn.Close()
}

func (c *portableConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *portableConn) RecvDHCP() (*Packet, *net.Interface, error) {
	var buf [1500]byte
	for {
		n, cm, _, err := c.conn.ReadFrom(buf[:])
		if err != nil {
			return nil, nil, err
		}
		pkt, err := Unmarshal(buf[:n])
		if err != nil {
			continue
		}
		intf, err := net.InterfaceByIndex(cm.IfIndex)
		if err != nil {
			return nil, nil, err
		}
		// TODO: possibly more validation that the source lines up
		// with what the packet.
		return pkt, intf, nil
	}
}

func (c *portableConn) SendDHCP(pkt *Packet, intf *net.Interface) error {
	b, err := pkt.Marshal()
	if err != nil {
		return err
	}

	switch pkt.TxType() {
	case TxBroadcast, TxHardwareAddr:
		cm := ipv4.ControlMessage{
			IfIndex: intf.Index,
		}
		addr := net.UDPAddr{
			IP:   net.IPv4bcast,
			Port: dhcpClientPort,
		}
		_, err = c.conn.WriteTo(b, &cm, &addr)
		return err
	case TxRelayAddr:
		// Send to the server port, not the client port.
		addr := net.UDPAddr{
			IP:   pkt.RelayAddr,
			Port: 67,
		}
		_, err = c.conn.WriteTo(b, nil, &addr)
		return err
	case TxClientAddr:
		addr := net.UDPAddr{
			IP:   pkt.ClientAddr,
			Port: dhcpClientPort,
		}
		_, err = c.conn.WriteTo(b, nil, &addr)
		return err
	default:
		return errors.New("unknown TX type for packet")
	}
}
