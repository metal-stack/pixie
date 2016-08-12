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
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// StaticBooter boots all machines with the same Spec.
//
// IDs in spec should be either local file paths, or HTTP/HTTPS URLs.
func StaticBooter(spec *Spec) Booter {
	ret := &staticBooter{
		kernel: string(spec.Kernel),
		spec: &Spec{
			Kernel:  "kernel",
			Cmdline: map[string]interface{}{},
			Message: spec.Message,
		},
	}
	for i, initrd := range spec.Initrd {
		ret.initrd = append(ret.initrd, string(initrd))
		ret.spec.Initrd = append(ret.spec.Initrd, ID(fmt.Sprintf("initrd-%d", i)))
	}
	for k, v := range spec.Cmdline {
		if id, ok := v.(ID); ok {
			ret.otherIDs = append(ret.otherIDs, string(id))
			ret.spec.Cmdline[k] = ID(fmt.Sprintf("other-%d", len(ret.otherIDs)-1))
		} else {
			ret.spec.Cmdline[k] = v
		}
	}

	return ret
}

type staticBooter struct {
	kernel   string
	initrd   []string
	otherIDs []string

	spec *Spec
}

func (s *staticBooter) BootSpec(m Machine) (*Spec, error) {
	return s.spec, nil
}

func (s *staticBooter) serveFile(path string) (io.ReadCloser, error) {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		resp, err := http.Get(path)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("%s: %s", path, http.StatusText(resp.StatusCode))
		}
		return resp.Body, nil
	}
	return os.Open(path)
}

func (s *staticBooter) ReadBootFile(id ID) (io.ReadCloser, error) {
	path := string(id)
	switch {
	case path == "kernel":
		return s.serveFile(s.kernel)

	case strings.HasPrefix(path, "initrd-"):
		i, err := strconv.Atoi(string(path[7:]))
		if err != nil || i < 0 || i >= len(s.initrd) {
			return nil, fmt.Errorf("no file with ID %q", id)
		}
		return s.serveFile(s.initrd[i])

	case strings.HasPrefix(path, "other-"):
		i, err := strconv.Atoi(string(path[6:]))
		if err != nil || i < 0 || i >= len(s.otherIDs) {
			return nil, fmt.Errorf("no file with ID %q", id)
		}
		return s.serveFile(s.otherIDs[i])
	}

	return nil, fmt.Errorf("no file with ID %q", id)
}

func (s *staticBooter) WriteBootFile(ID, io.Reader) error {
	return nil
}
