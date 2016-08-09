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
	"io/ioutil"
	"net"
	"strconv"

	"go.universe.tf/netboot/tftp"
)

func (s *Server) serveTFTP(l net.PacketConn) {
	ts := tftp.Server{
		Handler: s.handleTFTP,
		InfoLog: func(msg string) { fmt.Println(msg) },
	}
	err := ts.Serve(l)
	fmt.Printf("Serving TFTP failed: %s\n", err)
}

func (s *Server) handleTFTP(path string, clientAddr net.Addr) (io.ReadCloser, int64, error) {
	i, err := strconv.Atoi(path)
	if err != nil {
		return nil, 0, errors.New("not found")
	}

	bs, ok := s.Ipxe[Firmware(i)]
	if !ok {
		return nil, 0, errors.New("not found")
	}

	return ioutil.NopCloser(bytes.NewBuffer(bs)), int64(len(bs)), nil
}
