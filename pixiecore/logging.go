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
	"net/http"
)

func (s *Server) logf(format string, args ...interface{}) {
	if s.Log == nil {
		return
	}
	s.Log(fmt.Sprintf(format, args...))
}

// logHTTP logs a message with some context about the HTTP request
// that caused the statement to be logged.
func (s *Server) logHTTP(r *http.Request, format string, args ...interface{}) {
	if s.Log != nil {
		pfx := fmt.Sprintf("HTTP request for %s from %s: ", r.URL, r.RemoteAddr)
		s.logf(pfx+format, args...)
	}
}
