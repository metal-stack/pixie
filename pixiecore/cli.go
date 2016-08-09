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

package pixiecore

import (
	"fmt"
	"io"
	"net/http"

	"go.universe.tf/netboot/pixiecore/cmd"
)

type hellyeah struct{}

func (hellyeah) BootSpec(Machine) (*Spec, error) {
	return &Spec{
		Kernel: ID("k"),
		Initrd: []ID{"0", "1"},
	}, nil
}

func (hellyeah) ReadBootFile(p ID) (io.ReadCloser, error) {
	var url string
	switch p {
	case "k":
		url = "http://tinycorelinux.net/7.x/x86/release/distribution_files/vmlinuz64"
	case "0":
		url = "http://tinycorelinux.net/7.x/x86/release/distribution_files/rootfs.gz"
	case "1":
		url = "http://tinycorelinux.net/7.x/x86/release/distribution_files/modules64.gz"
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GET %q failed: %s", url, resp.Status)
	}
	return resp.Body, nil
}

func (hellyeah) WriteBootFile(ID, io.Reader) error {
	return nil
}

// CLI runs the Pixiecore commandline.
//
// Takes a map of ipxe bootloader binaries for various architectures.
func CLI(ipxe map[Firmware][]byte) {
	s := &Server{
		Booter: hellyeah{},
		Ipxe:   ipxe,
	}
	fmt.Println(s.Serve())

	cmd.Execute()
}
