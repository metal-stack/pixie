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

package cli

import (
	"fmt"
	"io"
	"net/http"

	"go.universe.tf/netboot/pixiecore"
)

// This is just a very temporary test booter that boots everything
// into tinycore linux, always.

type tinycore struct{}

func (tinycore) BootSpec(m pixiecore.Machine) (*pixiecore.Spec, error) {
	return &pixiecore.Spec{
		Kernel: pixiecore.ID("k"),
		Initrd: []pixiecore.ID{"1", "2"},
	}, nil
}

func (tinycore) ReadBootFile(id pixiecore.ID) (io.ReadCloser, error) {
	var url string
	switch id {
	case "k":
		url = "http://tinycorelinux.net/7.x/x86/release/distribution_files/vmlinuz64"
	case "1":
		url = "http://tinycorelinux.net/7.x/x86/release/distribution_files/rootfs.gz"
	case "2":
		url = "http://tinycorelinux.net/7.x/x86/release/distribution_files/modules64.gz"
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("%s: %s", url, http.StatusText(resp.StatusCode))
	}
	return resp.Body, nil
}

func (tinycore) WriteBootFile(id pixiecore.ID, body io.Reader) error {
	return nil
}
