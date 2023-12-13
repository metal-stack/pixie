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

package dhcp4

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"
)

const (
	// See: https://wiki.wireshark.org/Development/LibpcapFileFormat#global-header
	pcapHdrLen = 24
	ethHdrLen  = 14
	udpHdrLen  = 8
)

// See: https://wiki.wireshark.org/Development/LibpcapFileFormat#record-packet-header
type packetHdr struct {
	_   uint32
	_   uint32
	Len uint32
	_   uint32
}

func udpFromPcap(fname string) ([][]byte, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	r := bufio.NewReader(f)

	_, err = r.Discard(pcapHdrLen)
	if err != nil {
		return nil, fmt.Errorf("failed to discard header: %s", err)
	}

	var ret [][]byte
	for {
		var hdr packetHdr
		if err = binary.Read(r, binary.LittleEndian, &hdr); err != nil {
			break
		}
		bs := make([]byte, hdr.Len)
		if _, err = io.ReadFull(r, bs); err != nil {
			break
		}
		hdrLen := ethHdrLen
		hdrLen += int(bs[hdrLen]&0xF) * 4 // IP header
		hdrLen += udpHdrLen
		ret = append(ret, bs[hdrLen:])
	}
	if err != nil && err != io.EOF {
		return nil, err
	}

	return ret, nil
}

func TestParse(t *testing.T) {
	rawPkts, err := udpFromPcap("testdata/dhcp.pcap")
	if err != nil {
		t.Fatalf("Getting test packets from pcap: %s", err)
	}

	var pkts bytes.Buffer
	for i, rawPkt := range rawPkts {
		pkt, err := Unmarshal(rawPkt)
		if err != nil {
			t.Fatalf("Parsing DHCP packet #%d: %s", i+1, err)
		}
		pkts.WriteString(pkt.DebugString())
		pkts.WriteString("======\n")
	}

	expectedFile := "testdata/dhcp.parsed"
	expected, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("Reading expected file: %s", err)
	}

	if pkts.String() != string(expected) {
		if os.Getenv("UPDATE_TESTDATA") != "" {
			_ = os.WriteFile(expectedFile, pkts.Bytes(), 0644)
			t.Errorf("dhcp.pcap didn't decode to dhcp.parsed (updated dhcp.parsed)")
		} else {
			t.Fatalf("dhcp.pcap didn't decode to dhcp.parsed (rerun with UPDATE_TESTDATA=1 to get diff)")
		}
	}
}

func TestWriteRead(t *testing.T) {
	rawPkts, err := udpFromPcap("testdata/dhcp.pcap")
	if err != nil {
		t.Fatalf("Getting test packets from pcap: %s", err)
	}

	var pkts []*Packet
	for i, rawPkt := range rawPkts {
		pkt, err := Unmarshal(rawPkt)
		if err != nil {
			t.Fatalf("Unmarshalling testdata packet #%d: %s", i+1, err)
		}
		pkts = append(pkts, pkt)
	}

	for _, pkt := range pkts {
		raw, err := pkt.Marshal()
		if err != nil {
			t.Fatalf("Packet marshalling failed: %s\nPacket: %#v", err, pkt)
		}
		pkt2, err := Unmarshal(raw)
		if err != nil {
			t.Fatalf("Packet unmarshalling failed: %s\nPacket: %#v", err, pkt)
		}
		if !reflect.DeepEqual(pkt, pkt2) {
			t.Fatalf("Packet mutated by write-then-read: %#v", pkt)
		}
	}
}
