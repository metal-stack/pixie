package dhcp

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"go.universe.tf/netboot/pcap"
)

func udpFromPcap(fname string) ([][]byte, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	r, err := pcap.NewReader(f)
	if err != nil {
		return nil, err
	}

	if r.LinkType != pcap.LinkEthernet {
		return nil, errors.New("Pcap packets are not ethernet")
	}

	ret := [][]byte{}
	for r.Next() {
		// Assume here that the packets are UDPv4, and just chop off
		// the headers in front of the UDP payload
		pkt := r.Packet()
		hdrLen := 14                             // Ethernet header
		hdrLen += int(pkt.Bytes[hdrLen]&0xF) * 4 // IP header
		hdrLen += 8                              // UDP header
		ret = append(ret, pkt.Bytes[hdrLen:])
	}
	if r.Err() != nil {
		return nil, r.Err()
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
		pkts.WriteString(pkt.testString())
	}

	expectedFile := "testdata/dhcp.parsed"
	expected, err := ioutil.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("Reading expected file: %s", err)
	}

	if pkts.String() != string(expected) {
		if os.Getenv("UPDATE_TESTDATA") != "" {
			ioutil.WriteFile(expectedFile, pkts.Bytes(), 0644)
			t.Errorf("dhcp.pcap didn't decode to dhcp.parsed (updated dhcp.parsed)")
		} else {
			t.Fatalf("dhcp.pcap didn't decode to dhcp.parsed (rerun with UPDATE_TESTDATA=1 to get diff)")
		}
	}
}
