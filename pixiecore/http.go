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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"text/template"
	"time"
)

func serveHTTP(l net.Listener, handlers ...func(*http.ServeMux)) error {
	mux := http.NewServeMux()
	for _, h := range handlers {
		h(mux)
	}
	if err := http.Serve(l, mux); err != nil { // nolint:gosec
		return fmt.Errorf("HTTP server shut down: %w", err)
	}
	return nil
}

func (s *Server) serveHTTP(mux *http.ServeMux) {
	mux.HandleFunc("/_/ipxe", s.handleIpxe)
	mux.HandleFunc("/_/file", s.handleFile)
	mux.HandleFunc("/_/booting", s.handleBooting)
	mux.HandleFunc("/certs", s.handleCerts)
}

func (s *Server) handleIpxe(w http.ResponseWriter, r *http.Request) {
	overallStart := time.Now()
	macStr := r.URL.Query().Get("mac")
	if macStr == "" {
		s.Log.Debug("Bad request, missing MAC address", "url", r.URL, "remoteaddr", r.RemoteAddr)
		http.Error(w, "missing MAC address parameter", http.StatusBadRequest)
		return
	}
	archStr := r.URL.Query().Get("arch")
	if archStr == "" {
		s.Log.Debug("Bad request, missing architecture", "url", r.URL, "remoteaddr", r.RemoteAddr)
		http.Error(w, "missing architecture parameter", http.StatusBadRequest)
		return
	}

	mac, err := net.ParseMAC(macStr)
	if err != nil {
		s.Log.Debug("Bad request, invalid MAC address", "url", r.URL, "remoteaddr", r.RemoteAddr, "mac", macStr, "error", err)
		http.Error(w, "invalid MAC address", http.StatusBadRequest)
		return
	}

	i, err := strconv.Atoi(archStr)
	if err != nil {
		s.Log.Debug("Bad request, invalid architecture", "url", r.URL, "remoteaddr", r.RemoteAddr, "arch", archStr, "error", err)
		http.Error(w, "invalid architecture", http.StatusBadRequest)
		return
	}
	arch := Architecture(i)
	switch arch {
	case ArchIA32, ArchX64:
	default:
		s.Log.Debug("Bad request, unknown architecture", "url", r.URL, "remoteaddr", r.RemoteAddr, "arch", arch)
		http.Error(w, "unknown architecture", http.StatusBadRequest)
		return
	}

	mach := Machine{
		MAC:  mac,
		Arch: arch,
	}
	start := time.Now()
	spec, err := s.Booter.BootSpec(mach)
	s.Log.Debug("Get bootspec for", "mac", mac, "duration", time.Since(start))
	if err != nil {
		s.Log.Info("Couldn't get a bootspec for", "mac", mac, "url", r.URL, "remoteaddr", r.RemoteAddr, "error", err)
		http.Error(w, "couldn't get a bootspec", http.StatusInternalServerError)
		return
	}
	if spec == nil {
		// TODO: make ipxe abort netbooting so it can fall through to
		// other boot options - unsure if that's possible.
		s.Log.Debug("No boot spec for, ignoring boot request", "mac", mac, "url", r.URL, "remoteaddr", r.RemoteAddr)
		http.Error(w, "you don't netboot", http.StatusNotFound)
		return
	}
	start = time.Now()
	script, err := ipxeScript(mach, spec, r.Host)
	s.Log.Debug("Construct ipxe script for", "mac", mac, "duration", time.Since(start))
	if err != nil {
		s.Log.Info("Failed to assemble ipxe script for", "mac", mac, "url", r.URL, "remoteaddr", r.RemoteAddr, "error", err)
		http.Error(w, "couldn't get a boot script", http.StatusInternalServerError)
		return
	}

	s.Log.Info("Sending ipxe boot script", "remoteaddr", r.RemoteAddr)
	start = time.Now()
	s.machineEvent(mac, machineStateIpxeScript, "Sent iPXE boot script")
	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write(script)
	s.Log.Debug("Writing ipxe script to", "mac", mac, "duration", time.Since(start))
	s.Log.Debug("handleIpxe for", "mac", mac, "duration", time.Since(overallStart))
}

func (s *Server) handleFile(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		s.Log.Debug("Bad request, missing filename", "url", r.URL, "remoteaddr", r.RemoteAddr)
		http.Error(w, "missing filename", http.StatusBadRequest)
	}

	f, sz, err := s.Booter.ReadBootFile(ID(name))
	if err != nil {
		s.Log.Info("Error getting file", "name", name, "url", r.URL, "remoteaddr", r.RemoteAddr, "error", err)
		http.Error(w, "couldn't get file", http.StatusInternalServerError)
		return
	}
	defer func() {
		_ = f.Close()
	}()
	if sz >= 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(sz, 10))
	} else {
		s.Log.Info("Unknown file size, boot will be VERY slow (can your Booter provide file sizes?)", "name", name)
	}
	if _, err = io.Copy(w, f); err != nil {
		s.Log.Info("Copy failed", "name", name, "remoteaddr", r.RemoteAddr, "url", r.URL, "error", err)
		return
	}
	s.Log.Info("Sent file", "name", name, "remoteaddr", r.RemoteAddr)

	switch r.URL.Query().Get("type") {
	case "kernel":
		mac, err := net.ParseMAC(r.URL.Query().Get("mac"))
		if err != nil {
			s.Log.Info("File fetch provided invalid MAC address", "mac", r.URL.Query().Get("mac"))
			return
		}
		s.machineEvent(mac, machineStateKernel, "Sent kernel %q", name)
	case "initrd":
		mac, err := net.ParseMAC(r.URL.Query().Get("mac"))
		if err != nil {
			s.Log.Info("File fetch provided invalid MAC address", "mac", r.URL.Query().Get("mac"))
			return
		}
		s.machineEvent(mac, machineStateInitrd, "Sent initrd %q", name)
	}
}

func (s *Server) handleBooting(w http.ResponseWriter, r *http.Request) {
	// Return a no-op boot script, to satisfy iPXE. It won't get used,
	// the boot script deletes this image immediately after
	// downloading.
	_, _ = fmt.Fprintf(w, "# Booting")

	macStr := r.URL.Query().Get("mac")
	if macStr == "" {
		s.Log.Debug("Bad request, missing MAC address", "url", r.URL, "remoteaddr", r.RemoteAddr)
		return
	}
	mac, err := net.ParseMAC(macStr)
	if err != nil {
		s.Log.Debug("Bad request, invalid MAC address", "url", r.URL, "remoteaddr", r.RemoteAddr, "mac", macStr, "error", err)
		return
	}
	s.machineEvent(mac, machineStateBooted, "Booting into OS")
}

func ipxeScript(mach Machine, spec *Spec, serverHost string) ([]byte, error) {
	if spec.IpxeScript != "" {
		return []byte(spec.IpxeScript), nil
	}

	if spec.Kernel == "" {
		return nil, errors.New("spec is missing Kernel")
	}

	urlTemplate := fmt.Sprintf("http://%s/_/file?name=%%s&type=%%s&mac=%%s", serverHost)
	var b bytes.Buffer
	b.WriteString("#!ipxe\n")
	u := fmt.Sprintf(urlTemplate, url.QueryEscape(string(spec.Kernel)), "kernel", url.QueryEscape(mach.MAC.String()))
	fmt.Fprintf(&b, "kernel --name kernel %s\n", u)
	for i, initrd := range spec.Initrd {
		u = fmt.Sprintf(urlTemplate, url.QueryEscape(string(initrd)), "initrd", url.QueryEscape(mach.MAC.String()))
		fmt.Fprintf(&b, "initrd --name initrd%d %s\n", i, u)
	}

	fmt.Fprintf(&b, "imgfetch --name ready http://%s/_/booting?mac=%s ||\n", serverHost, url.QueryEscape(mach.MAC.String()))
	b.WriteString("imgfree ready ||\n")

	b.WriteString("boot kernel ")
	for i := range spec.Initrd {
		fmt.Fprintf(&b, "initrd=initrd%d ", i)
	}

	f := func(id string) string {
		return fmt.Sprintf("http://%s/_/file?name=%s", serverHost, url.QueryEscape(id))
	}
	cmdline, err := expandCmdline(spec.Cmdline, template.FuncMap{"ID": f})
	if err != nil {
		return nil, fmt.Errorf("expanding cmdline %q: %w", spec.Cmdline, err)
	}
	b.WriteString(cmdline)
	b.WriteByte('\n')

	return b.Bytes(), nil
}

func (s *Server) handleCerts(w http.ResponseWriter, r *http.Request) {
	js, err := json.MarshalIndent(s.MetalConfig, "", "  ")
	if err != nil {
		s.Log.Error("handleCerts unable to marshal grpc config", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.Log.Debug("handleCerts return grpc config")
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)
	if err != nil {
		s.Log.Error("handleCerts unable to write grpc config to response", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
