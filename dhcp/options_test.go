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
