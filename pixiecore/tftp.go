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
	"strconv"
	"strings"
	"time"

	"github.com/pin/tftp"
)

func (s *Server) serveTFTP(addr string) error {

	// use nil in place of handler to disable read or write operations
	tftpServer := tftp.NewServer(s.readHandler, nil)
	tftpServer.SetTimeout(time.Minute)     // optional
	err := tftpServer.ListenAndServe(addr) // blocks until s.Shutdown() is called
	if err != nil {
		return fmt.Errorf("TFTP server shut down: %w", err)
	}
	return nil
}

func extractInfo(path string) (net.HardwareAddr, int, error) {
	pathElements := strings.Split(path, "/")
	if len(pathElements) != 2 {
		return nil, 0, errors.New("not found")
	}

	mac, err := net.ParseMAC(pathElements[0])
	if err != nil {
		return nil, 0, fmt.Errorf("invalid MAC address %q", pathElements[0])
	}

	i, err := strconv.Atoi(pathElements[1])
	if err != nil {
		return nil, 0, errors.New("not found")
	}

	return mac, i, nil
}

// readHandler is called when client starts file download from server
func (s *Server) readHandler(path string, rf io.ReaderFrom) error {
	_, i, err := extractInfo(path)
	if err != nil {
		return fmt.Errorf("unknown path %q", path)
	}

	bs, ok := s.Ipxe[Firmware(i)]
	if !ok {
		return fmt.Errorf("unknown firmware type %d", i)
	}

	n, err := rf.ReadFrom(bytes.NewReader(bs))
	if err != nil {
		s.Log.Errorf("unable to send payload %s", err)
		return err
	}
	s.Log.Infof("sent %d bytes", n)
	return nil
}
