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

import "testing"

func TestOptionReading(t *testing.T) {
	o := Options{
		1: []byte{1, 2, 3},
		2: []byte{3},
		3: []byte{0, 1},
	}

	b, ok := o.Byte(2)
	if !ok {
		t.Fatalf("Option 2 should be a valid byte")
	}
	if b != 3 {
		t.Fatalf("Wanted value 3 for option 2, got %d", b)
	}

	b, ok = o.Byte(3)
	if ok {
		t.Fatalf("Option 3 shouldn't be a valid byte")
	}

	u, ok := o.Uint16(3)
	if !ok {
		t.Fatalf("Option 3 should be a valid byte")
	}
	if u != 1 {
		t.Fatalf("Wanted value 1 for option 3, got %d", u)
	}
}

func TestCopy(t *testing.T) {
	o := Options{
		1: []byte{2},
		2: []byte{3, 4},
	}

	o2 := o.Copy()
	delete(o2, 2)

	if len(o) != 2 {
		t.Fatalf("Mutating Option copy mutated the original")
	}
}
