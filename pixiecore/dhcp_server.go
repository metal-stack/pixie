// Copyright 2016 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
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

type dhcpServer struct {
	//dhcpConn *dhcp4.Conn // Used for sending only
	//icmpConn icmp.PacketConn
	leases []lease
}

const (
	leaseFree = iota
	leaseReserved
	leaseActive
)

type lease struct {
	state int
	//mac     net.HardwareAddr
	//expires time.Time
}

func newDHCPServer(address string, dhcpConn *dhcp4.Conn) (*dhcpServer, error) {
	var ret dhcpServer

	managedNet, err := findIPNet(address)
	if err != nil {
		return nil, err
	}

	for ip := managedNet.IP.Mask(managedNet.Mask); managedNet.Contains(ip); ip = nextIP(ip) {
		ret.leases = append(ret.leases, lease{})
		if ip.Equal(managedNet.IP) {
			ret.leases[len(ret.leases)-1].state = leaseActive // Active with no expiry == eternal
		}
	}
	return nil, errors.New("TODO")
}

func nextIP(ip net.IP) net.IP {
	ip = ip.To4()
	if ip == nil {
		return nil
	}
	ret := net.IPv4(ip[0], ip[1], ip[2], ip[3])
	for i := 3; i >= 0; i-- {
		ret[i]++
		if ret[i] != 0 {
			break
		}
	}
	return ret
}

func findIPNet(address string) (*net.IPNet, error) {
	wanted := net.ParseIP(address)
	if wanted == nil || wanted.To4() == nil {
		return nil, fmt.Errorf("Bad IPv4 address %q", address)
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		ipaddr, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		if !ipaddr.IP.Equal(wanted) {
			continue
		}
		return ipaddr, nil
	}

	return nil, fmt.Errorf("IP address %q not found on any network interface", address)
}
