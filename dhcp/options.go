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
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sort"
)

// Options stores DHCP options.
type Options map[int][]byte

// Unmarshal parses DHCP options into o.
func (o Options) Unmarshal(bs []byte) error {
	for len(bs) > 0 {
		opt := int(bs[0])
		switch opt {
		case 0:
			// Padding byte
			bs = bs[1:]
		case 255:
			// End of options
			return nil
		default:
			// In theory, DHCP permits multiple instances of the same
			// option in a packet, as a way to have option values >255
			// bytes. Unfortunately, this is very loosely specified as
			// "up to individual options", and AFAICT, isn't at all
			// used in the wild.
			//
			// So, for now, seeing the same option twice in a packet
			// is going to be an error, until I get a bug report about
			// something that actually does it.
			if _, ok := o[opt]; ok {
				return fmt.Errorf("packet has duplicate option %d (please file a bug with a pcap!)", opt)
			}
			if len(bs) < 2 {
				return fmt.Errorf("option %d has no length byte", opt)
			}
			l := int(bs[1])
			if len(bs[2:]) < l {
				return fmt.Errorf("option %d claims to have %d bytes of payload, but only has %d bytes", opt, l, len(bs[2:]))
			}
			o[opt] = bs[2 : 2+l]
			bs = bs[2+l:]
		}
	}

	return errors.New("options are not terminated by a 255 byte")
}

// Marshal returns the wire encoding of o.
func (o Options) Marshal() ([]byte, error) {
	var ret bytes.Buffer
	if err := o.MarshalTo(&ret); err != nil {
		return nil, err
	}
	return ret.Bytes(), nil
}

// MarshalTo serializes o into w.
func (o Options) MarshalTo(w io.Writer) error {
	opts, err := o.marshalLimited(w, 0, false)
	if err != nil {
		return err
	}
	if len(opts) > 0 {
		return errors.New("some options not written, but no limit was given (please file a bug)")
	}
	return nil
}

// Copy returns a shallow copy of o.
func (o Options) Copy() Options {
	ret := make(Options, len(o))
	for k, v := range o {
		ret[k] = v
	}
	return ret
}

// marshalLimited serializes o into w. If nBytes > 0, as many options
// as possible are packed into that many bytes, inserting padding as
// needed, and the remaining unwritten options are returned.
func (o Options) marshalLimited(w io.Writer, nBytes int, skip52 bool) (Options, error) {
	ks := make([]int, 0, len(o))
	for n := range o {
		if n <= 0 || n >= 255 {
			return nil, fmt.Errorf("invalid DHCP option number %d", n)
		}
		ks = append(ks, n)
	}
	sort.Ints(ks)

	ret := make(Options)
	for _, n := range ks {
		opt := o[n]
		if len(opt) > 255 {
			return nil, fmt.Errorf("DHCP option %d has value >255 bytes", n)
		}

		// If space is limited, verify that we can fit the option plus
		// the final end-of-options marker.
		if nBytes > 0 && ((skip52 && n == 52) || len(opt)+3 > nBytes) {
			ret[n] = opt
			continue
		}

		w.Write([]byte{byte(n), byte(len(opt))})
		w.Write(opt)
		nBytes -= len(opt) + 2
	}

	w.Write([]byte{255})
	nBytes--
	if nBytes > 0 {
		w.Write(make([]byte, nBytes))
	}

	return ret, nil
}

// Byte returns the value of single-byte option n, if the option value
// is indeed a single byte.
func (o Options) Byte(n int) (byte, bool) {
	v := o[n]
	if v == nil || len(v) != 1 {
		return 0, false
	}
	return v[0], true
}

func (o Options) Uint16(n int) (uint16, bool) {
	v := o[n]
	if v == nil || len(v) != 2 {
		return 0, false
	}
	return binary.BigEndian.Uint16(v[:2]), true
}
