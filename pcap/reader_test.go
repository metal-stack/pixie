package pcap

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/kr/pretty"
)

func TestFiles(t *testing.T) {
	for _, fname := range []string{"usec", "nsec"} {
		f, err := os.Open(fmt.Sprintf("testdata/%s.pcap", fname))
		if err != nil {
			t.Fatalf("Opening test input file: %s", err)
		}
		r, err := NewReader(f)
		if err != nil {
			t.Fatalf("Creating pcap reader: %s", err)
		}
		if r.LinkType != LinkEthernet {
			t.Errorf("Expected link type %d, got %d", LinkEthernet, r.LinkType)
		}
		pkts := []*Packet{}
	ReadLoop:
		for {
			pkt, err := r.Next()
			if err != nil {
				if err == io.EOF {
					break ReadLoop
				}
				t.Fatalf("Unexpected error reading packets: %s", err)
			}
			pkts = append(pkts, pkt)
		}
		res := pretty.Sprintf("%# v", pkts)
		expectedFile := fmt.Sprintf("testdata/%s.parsed", fname)
		expected, err := ioutil.ReadFile(expectedFile)
		if err != nil {
			t.Fatalf("Reading expected file: %s", err)
		}
		if res != string(expected) {
			if os.Getenv("UPDATE_TESTDATA") != "" {
				ioutil.WriteFile(expectedFile, []byte(res), 0644)
				t.Errorf("%s.pcap didn't decode to %s.parsed (updated %s.parsed)", fname, fname, fname)
			} else {
				t.Fatalf("%s.pcap didn't decode to %s.parsed (rerun with UPDATE_TESTDATA=1 to get diff)", fname, fname)
			}
		}
	}
}
