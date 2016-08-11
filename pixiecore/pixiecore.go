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
	"net"
	"net/http"

	"go.universe.tf/netboot/dhcp4"
)

const (
	portDHCP = 67
	portTFTP = 69
	portHTTP = 81
	portPXE  = 4011
)

// An ID is an identifier used by Booters to reference files.
type ID string

// Architecture describes a kind of CPU architecture.
type Architecture int

// Architecture types that Pixiecore knows how to boot.
//
// These architectures are self-reported by the booting machine. The
// machine may support additional execution modes. For example, legacy
// PC BIOS reports itself as an ArchIA32, but may also support ArchX64
// execution.
const (
	// ArchIA32 is a 32-bit x86 machine. It _may_ also support X64
	// execution, but Pixiecore has no way of knowing.
	ArchIA32 Architecture = iota
	// ArchX64 is a 64-bit x86 machine (aka amd64 aka X64).
	ArchX64
)

// A Machine describes a machine that is attempting to boot.
type Machine struct {
	MAC  net.HardwareAddr
	Arch Architecture
}

// A Spec describes a kernel and associated configuration.
type Spec struct {
	// The kernel to boot
	Kernel ID
	// Optional init ramdisks for linux kernels
	Initrd []ID
	// Optional kernel commandline arguments. Values can be strings,
	// numbers, or IDs. IDs get translated to an HTTP(S) URL.
	Cmdline map[string]interface{}
	// Message to print on the client machine before booting.
	Message string
}

// A Booter provides boot instructions and files for machines.
//
// Due to the stateless nature of various boot protocols, BootSpec()
// will be called multiple times in the course of a single boot
// attempt.
type Booter interface {
	// The given MAC address wants to know what it should
	// boot. What should Pixiecore make it boot?
	//
	// Returning an error or a nil BootSpec will make Pixiecore ignore
	// the client machine's request.
	BootSpec(m Machine) (*Spec, error)
	// Get the bytes corresponding to an ID given in Spec.
	ReadBootFile(id ID) (io.ReadCloser, error)
	// Write the given Reader to an ID given in Spec.
	WriteBootFile(id ID, body io.Reader) error
}

// Firmware describes a kind of firmware attempting to boot.
//
// This should only be used for selecting the right bootloader within
// Pixiecore, kernel selection should key off the more generic
// Architecture.
type Firmware int

// The bootloaders that Pixiecore knows how to handle.
const (
	// Note the values match the values from RFC4578.
	FirmwareX86PC Firmware = 0 // "Classic" x86 BIOS with PXE/UNDI support.
	FirmwareEFI32          = 6 // 32-bit x86 processor running EFI
	FirmwareEFI64          = 7 // 64-bit x86 processor running EFI
)

var fwToArch = map[Firmware]Architecture{
	FirmwareX86PC: ArchIA32,
	FirmwareEFI32: ArchIA32,
	FirmwareEFI64: ArchX64,
}

// A Server boots machines using a Booter.
type Server struct {
	Booter Booter

	// Address to listen on, or empty for all interfaces.
	Address  string
	HTTPPort int

	// Ipxe lists the supported bootable Firmwares, and their
	// associated ipxe binary.
	Ipxe map[Firmware][]byte

	// Log receives logs on Pixiecore's operation. If nil, logging
	// is suppressed.
	//
	// TODO: until Trace is better defined and plumbed through, we use
	// Log extensively, which may make it too noisy for normal
	// use. Need to decide what, if anything, we should do about that.
	Log func(msg string)

	// These ports can technically be set for testing, but the
	// protocols burned in firmware on the client side hardcode these,
	// so if you change them in production, nothing will work.
	DHCPPort int
	TFTPPort int
	PXEPort  int

	// Trace receives huge amounts of detail about what Pixiecore is
	// doing, including raw packets it sent/received. This should be
	// nil unless you intend to provide a bug report.
	//
	// TODO: figure out the format this logs in, and write a
	// decoder. It'll likely include bits of pcap for packets
	// received, interleaved with what bits of Pixiecore thought about
	// them, so that you get a timeline complete with wire traffic.
	Trace io.Writer
}

// Serve listens for machines attempting to boot, and uses Booter to
// help them.
func (s *Server) Serve() error {
	if s.DHCPPort == 0 {
		s.DHCPPort = portDHCP
	}
	if s.TFTPPort == 0 {
		s.TFTPPort = portTFTP
	}
	if s.PXEPort == 0 {
		s.PXEPort = portPXE
	}
	if s.HTTPPort == 0 {
		s.HTTPPort = portHTTP
	}

	dhcp, err := dhcp4.NewConn(fmt.Sprintf("%s:%d", s.Address, s.DHCPPort))
	if err != nil {
		return err
	}
	tftp, err := net.ListenPacket("udp", fmt.Sprintf("%s:%d", s.Address, s.TFTPPort))
	if err != nil {
		dhcp.Close()
		return err
	}
	pxe, err := net.ListenPacket("udp", fmt.Sprintf("%s:%d", s.Address, s.PXEPort))
	if err != nil {
		dhcp.Close()
		tftp.Close()
		return err
	}

	// TODO: have something here for orderly shutdown when things go wrong.

	go s.serveDHCP(dhcp)
	go s.servePXE(pxe)
	go s.serveTFTP(tftp)
	http.HandleFunc("/_/ipxe", s.handleIpxe)
	http.HandleFunc("/_/file", s.handleFile)
	http.ListenAndServe(fmt.Sprintf("%s:%d", s.Address, s.HTTPPort), nil)

	return nil
}
