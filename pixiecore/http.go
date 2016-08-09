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

func (s *Server) handleIpxe(w http.ResponseWriter, r *http.Request) {
	args := r.URL.Query()
	mac, err := net.ParseMAC(args.Get("mac"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid MAC address %q: %s\n", args.Get("mac"), err), http.StatusBadRequest)
		return
	}

	i, err := strconv.Atoi(args.Get("arch"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid architecture %q: %s\n", args.Get("arch"), err), http.StatusBadRequest)
		return

	}
	arch := Architecture(i)
	switch arch {
	case ArchIA32, ArchX64:
	default:
		http.Error(w, fmt.Sprintf("Unknown architecture %q\n", arch), http.StatusBadRequest)
		return
	}

	mach := Machine{
		MAC:  mac,
		Arch: arch,
	}
	spec, err := s.Booter.BootSpec(mach)
	if err != nil {
		// TODO: maybe don't send this error over the network?
		http.Error(w, fmt.Sprintf("Error getting bootspec for %#v: %s\n", mach, err), http.StatusInternalServerError)
		return
	}
	if spec == nil {
		// TODO: consider making ipxe abort netbooting so it can fall
		// through to other boot options - unsure if that's possible.
		http.Error(w, fmt.Sprintf("Should not boot %q\n", mach.MAC), http.StatusServiceUnavailable)
		return
	}
	if spec.Kernel == "" {
		// TODO: maybe don't send this error over the network?
		http.Error(w, fmt.Sprintf("Invalid bootspec for %q: missing kernel\n", mach.MAC), http.StatusInternalServerError)
		return
	}

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
			// TODO: maybe don't send this error over the network?
			http.Error(w, fmt.Sprintf("Invalid bootspec for %q: unknown cmdline type\n", mach.MAC), http.StatusInternalServerError)
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
		// TODO: maybe don't send this error over the network?
		http.Error(w, fmt.Sprintf("%s\n", err), http.StatusInternalServerError)
		return
	}
	defer f.Close()
	if _, err = io.Copy(w, f); err != nil {
		// TODO: logger
		fmt.Println("Copy failed:", err)
	}
}
