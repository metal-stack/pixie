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
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

func (s *Server) httpError(w http.ResponseWriter, r *http.Request, status int, format string, args ...interface{}) {
	s.logHTTP(r, format, args...)
	http.Error(w, fmt.Sprintf(format, args...), status)
}

func (s *Server) handleIpxe(w http.ResponseWriter, r *http.Request) {
	args := r.URL.Query()
	mac, err := net.ParseMAC(args.Get("mac"))
	if err != nil {
		s.httpError(w, r, http.StatusBadRequest, "invalid MAC address %q: %s\n", args.Get("mac"), err)
		return
	}

	i, err := strconv.Atoi(args.Get("arch"))
	if err != nil {
		s.httpError(w, r, http.StatusBadRequest, "invalid architecture %q: %s\n", args.Get("arch"), err)
		return

	}
	arch := Architecture(i)
	switch arch {
	case ArchIA32, ArchX64:
	default:
		s.httpError(w, r, http.StatusBadRequest, "Unknown architecture %q\n", arch)
		return
	}

	mach := Machine{
		MAC:  mac,
		Arch: arch,
	}
	spec, err := s.Booter.BootSpec(mach)
	if err != nil {
		// TODO: maybe don't send this error over the network?
		s.logHTTP(r, "error getting bootspec for %#v: %s", mach, err)
		http.Error(w, "couldn't get a bootspec", http.StatusInternalServerError)
		return
	}
	if spec == nil {
		// TODO: make ipxe abort netbooting so it can fall through to
		// other boot options - unsure if that's possible.
		s.httpError(w, r, http.StatusNotFound, "no bootspec found for %q", mach.MAC)
		return
	}
	if spec.Kernel == "" {
		// TODO: maybe don't send this error over the network?
		s.logHTTP(r, "invalid bootspec for %q: missing kernel", mach.MAC)
		http.Error(w, "couldn't get a bootspec", http.StatusInternalServerError)
		return
	}

	// All is well, assemble the iPXE script.

	urlPrefix := fmt.Sprintf("http://%s/_/file?name=", r.Host)

	var b bytes.Buffer
	b.WriteString("#!ipxe\n")
	fmt.Fprintf(&b, "kernel --name kernel %s%s\n", urlPrefix, url.QueryEscape(string(spec.Kernel)))
	for i, initrd := range spec.Initrd {
		fmt.Fprintf(&b, "initrd --name initrd%d %s%s\n", i, urlPrefix, url.QueryEscape(string(initrd)))
	}
	b.WriteString("boot kernel ")
	for i := range spec.Initrd {
		fmt.Fprintf(&b, "initrd=initrd%d ", i)
	}
	for k, v := range spec.Cmdline {
		switch val := v.(type) {
		case string:
			fmt.Fprintf(&b, "%s=%s ", k, val)
		case int, int32, int64, uint32, uint64:
			fmt.Fprintf(&b, "%s=%d ", k, v)
		case ID:
			fmt.Fprintf(&b, "%s=%s%s ", k, urlPrefix, url.QueryEscape(string(val)))
		default:
			s.logHTTP(r, "invalid bootspec for %q: unknown cmdline type for key %q", mach.MAC, k)
			http.Error(w, "couldn't get a bootspec", http.StatusInternalServerError)
			return
		}
	}
	b.WriteByte('\n')
	w.Header().Set("Content-Type", "text/plain")
	b.WriteTo(w)
}

func (s *Server) handleFile(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	f, err := s.Booter.ReadBootFile(ID(name))
	if err != nil {
		s.logHTTP(r, "error getting requested file %q: %s", name, err)
		http.Error(w, "couldn't get file", http.StatusInternalServerError)
		return
	}
	defer f.Close()
	if _, err = io.Copy(w, f); err != nil {
		s.logHTTP(r, "copy of file %q failed: %s", name, err)
	}
}
