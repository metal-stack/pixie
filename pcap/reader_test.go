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

package pcap

import (
	"fmt"
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
		for r.Next() {
			pkts = append(pkts, r.Packet())
		}
		if r.Err() != nil {
			t.Fatalf("Reading packets from %s.pcap: %s", fname, r.Err())
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
