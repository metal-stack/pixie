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
		Handler:     s.handleTFTP,
		InfoLog:     func(msg string) { s.debug("TFTP", msg) },
		TransferLog: s.logTFTPTransfer,
	}
	err := ts.Serve(l)
	if err != nil {
		// TODO: fatal errors that return from one of the handler
		// goroutines should plumb the error back to the
		// coordinating goroutine, so that it can do an orderly
		// shutdown and return the error from Serve(). This "log +
		// randomly stop a piece of pixiecore" is a terrible
		// kludge.
		s.log("TFTP", "Server shut down unexpectedly: %s", err)
	}
}

func (s *Server) logTFTPTransfer(clientAddr net.Addr, path string, err error) {
	if err != nil {
		s.log("TFTP", "Send of %q to %s failed: %s", path, clientAddr, err)
	} else {
		s.log("TFTP", "Sent %q to %s", path, clientAddr)
	}
}

func (s *Server) handleTFTP(path string, clientAddr net.Addr) (io.ReadCloser, int64, error) {
	i, err := strconv.Atoi(path)
	if err != nil {
		return nil, 0, errors.New("not found")
	}

	bs, ok := s.Ipxe[Firmware(i)]
	if !ok {
		return nil, 0, fmt.Errorf("unknown firmware type %d", i)
	}

	return ioutil.NopCloser(bytes.NewBuffer(bs)), int64(len(bs)), nil
}
