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

func (s *Server) serveTFTP(l net.PacketConn) error {

	// use nil in place of handler to disable read or write operations
	tftpServer := tftp.NewServer(s.readHandler, nil)
	tftpServer.SetTimeout(time.Minute)      // optional
	err := tftpServer.ListenAndServe(":69") // blocks until s.Shutdown() is called

	// ts := tftp.Server{
	// 	Handler:     s.handleTFTP,
	// 	InfoLog:     func(msg string) { s.Log.Debugf("TFTP", msg) },
	// 	TransferLog: s.logTFTPTransfer,
	// }
	// err := ts.Serve(l)
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

// func (s *Server) logTFTPTransfer(clientAddr net.Addr, path string, err error) {
// 	mac, _, pathErr := extractInfo(path)
// 	if pathErr != nil {
// 		s.Log.Infof("unable to extract mac from request:%v", pathErr)
// 		return
// 	}
// 	if err != nil {
// 		s.Log.Infof("Send of %q to %s failed: %s", path, clientAddr, err)
// 	} else {
// 		s.Log.Infof("Sent %q to %s", path, clientAddr)
// 		s.machineEvent(mac, machineStateTFTP, "Sent iPXE to %s", clientAddr)
// 	}
// }

func (s *Server) handleTFTP(path string, clientAddr net.Addr) (io.ReadCloser, int64, error) {
	_, i, err := extractInfo(path)
	if err != nil {
		return nil, 0, fmt.Errorf("unknown path %q", path)
	}

	bs, ok := s.Ipxe[Firmware(i)]
	if !ok {
		return nil, 0, fmt.Errorf("unknown firmware type %d", i)
	}

	return io.NopCloser(bytes.NewBuffer(bs)), int64(len(bs)), nil
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
		s.Log.Errorf("TFTP", "unable to send payload %s", err)
		return err
	}
	s.Log.Infof("TFTP", "sent %d", n)
	return nil
}
