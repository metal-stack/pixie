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

package tftp

import (
	"bytes"
	"io"
	"net"
)

// ConstantHandler returns a Handler that serves bs for all requested paths.
func ConstantHandler(bs []byte) Handler {
	return func(path string, addr net.Addr) (io.ReadCloser, int64, error) {
		return io.NopCloser(bytes.NewBuffer(bs)), int64(len(bs)), nil
	}
}
