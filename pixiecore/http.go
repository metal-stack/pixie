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
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"text/template"
)

func (s *Server) serveHTTP(l net.Listener) error {
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/_/ipxe", s.handleIpxe)
	httpMux.HandleFunc("/_/file", s.handleFile)
	if err := http.Serve(l, httpMux); err != nil {
		return fmt.Errorf("HTTP server shut down: %s", err)
	}
	return nil
}

func (s *Server) handleIpxe(w http.ResponseWriter, r *http.Request) {
	macStr := r.URL.Query().Get("mac")
	if macStr == "" {
		s.debug("HTTP", "Bad request %q from %s, missing MAC address", r.URL, r.RemoteAddr)
		http.Error(w, "missing MAC address parameter", http.StatusBadRequest)
		return
	}
	archStr := r.URL.Query().Get("arch")
	if archStr == "" {
		s.debug("HTTP", "Bad request %q from %s, missing architecture", r.URL, r.RemoteAddr)
		http.Error(w, "missing architecture parameter", http.StatusBadRequest)
		return
	}

	mac, err := net.ParseMAC(macStr)
	if err != nil {
		s.debug("HTTP", "Bad request %q from %s, invalid MAC address %q (%s)", r.URL, r.RemoteAddr, macStr, err)
		http.Error(w, "invalid MAC address", http.StatusBadRequest)
		return
	}

	i, err := strconv.Atoi(archStr)
	if err != nil {
		s.debug("HTTP", "Bad request %q from %s, invalid architecture %q (%s)", r.URL, r.RemoteAddr, archStr, err)
		http.Error(w, "invalid architecture", http.StatusBadRequest)
		return
	}
	arch := Architecture(i)
	switch arch {
	case ArchIA32, ArchX64:
	default:
		s.debug("HTTP", "Bad request %q from %s, unknown architecture %q", r.URL, r.RemoteAddr, arch)
		http.Error(w, "unknown architecture", http.StatusBadRequest)
		return
	}

	mach := Machine{
		MAC:  mac,
		Arch: arch,
	}
	spec, err := s.Booter.BootSpec(mach)
	if err != nil {
		s.log("HTTP", "Couldn't get a bootspec for %s (query %q from %s): %s", mac, r.URL, r.RemoteAddr, err)
		http.Error(w, "couldn't get a bootspec", http.StatusInternalServerError)
		return
	}
	if spec == nil {
		// TODO: make ipxe abort netbooting so it can fall through to
		// other boot options - unsure if that's possible.
		s.debug("HTTP", "No boot spec for %s (query %q from %s), ignoring boot request", mac, r.URL, r.RemoteAddr)
		http.Error(w, "you don't netboot", http.StatusNotFound)
		return
	}
	script, err := ipxeScript(spec, r.Host)
	if err != nil {
		s.log("HTTP", "Failed to assemble ipxe script for %s (query %q from %s): %s", mac, r.URL, r.RemoteAddr, err)
		http.Error(w, "couldn't get a boot script", http.StatusInternalServerError)
		return
	}

	s.log("HTTP", "Sending ipxe boot script to %s", r.RemoteAddr)
	w.Header().Set("Content-Type", "text/plain")
	w.Write(script)
}

func (s *Server) handleFile(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		s.debug("HTTP", "Bad request %q from %s, missing filename", r.URL, r.RemoteAddr)
		http.Error(w, "missing filename", http.StatusBadRequest)
	}

	f, sz, err := s.Booter.ReadBootFile(ID(name))
	if err != nil {
		s.log("HTTP", "Error getting file %q (query %q from %s): %s", name, r.URL, r.RemoteAddr, err)
		http.Error(w, "couldn't get file", http.StatusInternalServerError)
		return
	}
	defer f.Close()
	if sz >= 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(sz, 10))
	} else {
		s.log("HTTP", "Unknown file size for %q, boot will be VERY slow (can your Booter provide file sizes?")
	}
	if _, err = io.Copy(w, f); err != nil {
		s.log("HTTP", "Copy of %q to %s (query %q) failed: %s", name, r.RemoteAddr, r.URL, err)
		return
	}
	s.log("HTTP", "Sent file %q to %s", name, r.RemoteAddr)
}

func ipxeScript(spec *Spec, serverHost string) ([]byte, error) {
	if spec.Kernel == "" {
		return nil, errors.New("spec is missing Kernel")
	}

	urlPrefix := fmt.Sprintf("http://%s/_/file?name=", serverHost)
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

	f := func(id string) string {
		return fmt.Sprintf("http://%s/_/file?name=%s", serverHost, url.QueryEscape(id))
	}
	cmdline, err := expandCmdline(spec.Cmdline, template.FuncMap{"ID": f})
	if err != nil {
		return nil, fmt.Errorf("expanding cmdline %q: %s", spec.Cmdline, err)
	}
	b.WriteString(cmdline)

	b.WriteByte('\n')
	return b.Bytes(), nil
}
