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
	"html/template"
	"io/ioutil"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/ui"
)

//go:generate go-bindata -o ui/autogen.go -ignore autogen.go -pkg ui -nometadata -nomemcopy -prefix ui/ ui/

const assetsPath = "/_/assets/"

func (s *Server) serveUI(mux *http.ServeMux) {
	mux.HandleFunc("/", s.handleUI)
	mux.HandleFunc(assetsPath, s.handleUI)
}

func (s *Server) handleUI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		r.URL.Path = assetsPath + "index.html"
	}
	if !strings.HasPrefix(r.URL.Path, assetsPath) {
		http.NotFound(w, r)
		return
	}
	// Sticking a / in front of the path before cleaning it will strip
	// out any "../" attempts at path traversal. Then we remove the
	// leading /, and we end up with a path we can filepath.Join or
	// fetch from asset data.
	path := filepath.Clean("/" + r.URL.Path[len(assetsPath):])[1:]
	t, err := s.getTemplate(path)
	if err != nil {
		s.log("UI", "Failed to parse template for %q: %s", path, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	var b bytes.Buffer
	s.eventsMu.Lock()
	defer s.eventsMu.Unlock()
	if err = t.Execute(&b, s.events); err != nil {
		s.log("UI", "Failed to expand template for %q: %s", path, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	mimetype := mime.TypeByExtension(filepath.Ext(path))
	if mimetype != "" {
		w.Header().Set("Content-Type", mimetype)
	}

	b.WriteTo(w)
}

func (s *Server) getTemplate(name string) (*template.Template, error) {
	var (
		bs  []byte
		err error
	)
	if s.UIAssetsDir != "" {
		bs, err = ioutil.ReadFile(filepath.Join(s.UIAssetsDir, name))
	} else {
		bs, err = ui.Asset(name)
	}
	if err != nil {
		return nil, err
	}
	funcs := template.FuncMap{
		"dec": func(i int) int {
			return i - 1
		},
		"timestamp_millis": func(t time.Time) int64 {
			return t.UnixNano() / int64(time.Millisecond)
		},
	}
	return template.New(name).Funcs(funcs).Parse(string(bs))
}
