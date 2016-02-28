package pcap

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"testing"
	"time"

	"github.com/kr/pretty"
)

func TestReadback(t *testing.T) {
	pkts := []*Packet{
		{
			Timestamp: time.Now(),
			Length:    42,
			Bytes:     []byte{1, 2, 3, 4},
		},
		{
			Timestamp: time.Now(),
			Length:    30,
			Bytes:     []byte{2, 3, 4, 5},
		},
		{
			Timestamp: time.Now(),
			Length:    20,
			Bytes:     []byte{3, 4, 5, 6},
		},
		{
			Timestamp: time.Now(),
			Length:    10,
			Bytes:     []byte{4, 5, 6, 7},
		},
	}

	serializations := map[string]bool{}
	for _, order := range []binary.ByteOrder{nil, binary.LittleEndian, binary.BigEndian} {
		var b bytes.Buffer
		w := &Writer{
			Writer:    &b,
			LinkType:  LinkEthernet,
			SnapLen:   65535,
			ByteOrder: order,
		}

		for _, pkt := range pkts {
			if err := w.Put(pkt); err != nil {
				t.Fatalf(pretty.Sprintf("Writing packet %# v: %s", pkt, err))
			}
		}

		// Record the binary form, to check how many different serializations we get.
		serializations[b.String()] = true

		readBack := []*Packet{}
		r, err := NewReader(&b)
		if err != nil {
			t.Fatalf("Initializing reader for writer read-back: %s", err)
		}
		if r.LinkType != LinkEthernet {
			t.Fatalf("Wrote link type %d, read back %d", LinkEthernet, r.LinkType)
		}

		for r.Next() {
			readBack = append(readBack, r.Packet())
		}
		if r.Err() != nil {
			t.Fatalf("Reading packets back: %s", r.Err())
		}

		if !reflect.DeepEqual(pkts, readBack) {
			t.Fatalf("Packets were mutated by write-then-read")
		}
	}

	if len(serializations) != 2 {
		t.Fatalf("Expected 2 distinct serializations due to endianness, got %d", len(serializations))
	}
}
