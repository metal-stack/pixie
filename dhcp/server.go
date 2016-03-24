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

package dhcp // import "go.universe.tf/netboot/dhcp"

import (
	"io"
	"net"
)

// TxType describes how a Packet should be sent on the wire.
type TxType int

// The various transmission strategies described in RFC 2131. "MUST",
// "MUST NOT", "SHOULD" and "MAY" are as specified in RFC 2119.
const (
	// Packet MUST be broadcast.
	TxBroadcast TxType = iota
	// Packet MUST be unicasted to port 67 of RelayAddr
	TxRelayAddr
	// Packet MUST be unicasted to port 68 of ClientAddr
	TxClientAddr
	// Packet SHOULD be unicasted to port 68 of YourAddr, with the
	// link-layer destination explicitly set to HardwareAddr. You MUST
	// NOT rely on ARP resolution to discover the link-layer
	// destination address.
	//
	// Conn implementations that cannot explicitly set the link-layer
	// destination address MAY instead broadcast the packet.
	TxHardwareAddr
)

// Conn is a DHCP-oriented packet socket.
//
// Multiple goroutines may invoke methods on a Conn simultaneously.
type Conn interface {
	io.Closer
	// RecvDHCP reads a Packet from the connection. It returns the
	// packet and the interface it was received on, which may be nil
	// if interface information cannot be obtained.
	RecvDHCP() (pkt *Packet, intf *net.Interface, err error)
	// SendDHCP sends pkt. The precise transmission mechanism depends
	// on pkt.TxType(). intf should be the net.Interface returned by
	// RecvDHCP if responding to a DHCP client, or the interface for
	// which configuration is desired if acting as a client.
	SendDHCP(pkt *Packet, intf *net.Interface) error
}
