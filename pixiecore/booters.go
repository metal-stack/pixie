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
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/nacl/secretbox"
)

// StaticBooter boots all machines with the same Spec.
//
// IDs in spec should be either local file paths, or HTTP/HTTPS URLs.
func StaticBooter(spec *Spec) Booter {
	ret := &staticBooter{
		kernel: string(spec.Kernel),
		spec: &Spec{
			Kernel:  "kernel",
			Cmdline: map[string]interface{}{},
			Message: spec.Message,
		},
	}
	for i, initrd := range spec.Initrd {
		ret.initrd = append(ret.initrd, string(initrd))
		ret.spec.Initrd = append(ret.spec.Initrd, ID(fmt.Sprintf("initrd-%d", i)))
	}
	for k, v := range spec.Cmdline {
		if id, ok := v.(ID); ok {
			ret.otherIDs = append(ret.otherIDs, string(id))
			ret.spec.Cmdline[k] = ID(fmt.Sprintf("other-%d", len(ret.otherIDs)-1))
		} else {
			ret.spec.Cmdline[k] = v
		}
	}

	return ret
}

type staticBooter struct {
	kernel   string
	initrd   []string
	otherIDs []string

	spec *Spec
}

func (s *staticBooter) BootSpec(m Machine) (*Spec, error) {
	return s.spec, nil
}

func (s *staticBooter) serveFile(path string) (io.ReadCloser, error) {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		resp, err := http.Get(path)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("%s: %s", path, http.StatusText(resp.StatusCode))
		}
		return resp.Body, nil
	}
	return os.Open(path)
}

func (s *staticBooter) ReadBootFile(id ID) (io.ReadCloser, error) {
	path := string(id)
	switch {
	case path == "kernel":
		return s.serveFile(s.kernel)

	case strings.HasPrefix(path, "initrd-"):
		i, err := strconv.Atoi(string(path[7:]))
		if err != nil || i < 0 || i >= len(s.initrd) {
			return nil, fmt.Errorf("no file with ID %q", id)
		}
		return s.serveFile(s.initrd[i])

	case strings.HasPrefix(path, "other-"):
		i, err := strconv.Atoi(string(path[6:]))
		if err != nil || i < 0 || i >= len(s.otherIDs) {
			return nil, fmt.Errorf("no file with ID %q", id)
		}
		return s.serveFile(s.otherIDs[i])
	}

	return nil, fmt.Errorf("no file with ID %q", id)
}

func (s *staticBooter) WriteBootFile(ID, io.Reader) error {
	return nil
}

// APIBooter gets a BootSpec from a remote server over HTTP.
//
// The API is described in README.api.md
func APIBooter(url string, timeout time.Duration) (Booter, error) {
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	ret := &apibooter{
		client:    &http.Client{Timeout: timeout},
		urlPrefix: url + "v1",
	}
	if _, err := io.ReadFull(rand.Reader, ret.key[:]); err != nil {
		return nil, fmt.Errorf("failed to get randomness for signing key: %s", err)
	}

	return ret, nil
}

type apibooter struct {
	client    *http.Client
	urlPrefix string
	key       [32]byte
}

func (b *apibooter) getAPIResponse(hw net.HardwareAddr) (io.ReadCloser, error) {
	reqURL := fmt.Sprintf("%s/boot/%s", b.urlPrefix, hw)
	resp, err := b.client.Get(reqURL)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("%s: %s", reqURL, http.StatusText(resp.StatusCode))
	}

	return resp.Body, nil
}

func (b *apibooter) BootSpec(m Machine) (*Spec, error) {
	body, err := b.getAPIResponse(m.MAC)
	defer body.Close()
	if err != nil {
		return nil, err
	}

	r := struct {
		Kernel  string      `json:"kernel"`
		Initrd  []string    `json:"initrd"`
		Cmdline interface{} `json:"cmdline"`
		Message string      `json:"message"`
	}{}

	if err = json.NewDecoder(body).Decode(&r); err != nil {
		return nil, err
	}

	r.Kernel, err = b.makeURLAbsolute(r.Kernel)
	if err != nil {
		return nil, err
	}
	for i, img := range r.Initrd {
		r.Initrd[i], err = b.makeURLAbsolute(img)
		if err != nil {
			return nil, err
		}
	}

	ret := Spec{
		Message: r.Message,
	}
	if ret.Kernel, err = b.signURL(r.Kernel); err != nil {
		return nil, err
	}
	for _, img := range r.Initrd {
		initrd, err := b.signURL(img)
		if err != nil {
			return nil, err
		}
		ret.Initrd = append(ret.Initrd, initrd)
	}

	if r.Cmdline != nil {
		switch c := r.Cmdline.(type) {
		case string:
			ret.Cmdline[c] = ""
		case map[string]interface{}:
			ret.Cmdline, err = b.constructCmdline(c)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("API server returned unknown type %T for kernel cmdline", r.Cmdline)
		}
	}

	return &ret, nil
}

func (b *apibooter) ReadBootFile(id ID) (io.ReadCloser, error) {
	urlStr, err := b.getURL(id)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("%q is not an URL", urlStr)
	}
	var ret io.ReadCloser
	if u.Scheme == "file" {
		ret, err = b.readLocal(u)
	} else {
		// urlStr will get reparsed by http.Get, which is mildly
		// wasteful, but the code looks nicer than constructing a
		// Request.
		ret, err = b.readRemote(urlStr)
	}
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (b *apibooter) WriteBootFile(id ID, body io.Reader) error {
	u, err := b.getURL(id)
	if err != nil {
		return err
	}

	resp, err := http.Post(u, "application/octet-stream", body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("POST %q failed: %s", u, resp.Status)
	}
	defer resp.Body.Close()
	return nil
}

func (b *apibooter) makeURLAbsolute(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("%q is not an URL", urlStr)
	}
	if !u.IsAbs() {
		base, err := url.Parse(b.urlPrefix)
		if err != nil {
			return "", err
		}
		u = base.ResolveReference(u)
	}
	return u.String(), nil
}

func (b *apibooter) constructCmdline(m map[string]interface{}) (map[string]interface{}, error) {
	ret := map[string]interface{}{}
	for k, vi := range m {
		switch v := vi.(type) {
		case bool:
			ret[k] = ""
		case string:
			ret[k] = v
		case map[string]interface{}:
			urlStr, ok := v["url"].(string)
			if !ok {
				return nil, fmt.Errorf("cmdline key %q has object value with no 'url' attribute", k)
			}
			urlStr, err := b.makeURLAbsolute(urlStr)
			if err != nil {
				return nil, fmt.Errorf("invalid url for cmdline key %q: %s", k, err)
			}
			encoded, err := b.signURL(urlStr)
			if err != nil {
				return nil, err
			}
			ret[k] = encoded
		default:
			return nil, fmt.Errorf("unsupported value kind %T for cmdline key %q", vi, k)
		}
	}
	return ret, nil
}

func (b *apibooter) signURL(u string) (ID, error) {
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return "", fmt.Errorf("could not read randomness for signing nonce: %s", err)
	}

	out := nonce[:]

	// Secretbox is authenticated encryption. In theory we only need
	// symmetric authentication, but secretbox is stupidly simple to
	// use and hard to get wrong, and the encryption overhead should
	// be tiny for such a small URL unless you're trying to
	// simultaneously netboot a million machines. This is one case
	// where convenience and certainty that you got it right trumps
	// pure efficiency.
	out = secretbox.Seal(out, []byte(u), &nonce, &b.key)
	return ID(base64.URLEncoding.EncodeToString(out)), nil
}

func (b *apibooter) getURL(signedStr ID) (string, error) {
	signed, err := base64.URLEncoding.DecodeString(string(signedStr))
	if err != nil {
		return "", err
	}
	if len(signed) < 24 {
		return "", errors.New("signed blob too short to be valid")
	}

	var nonce [24]byte
	copy(nonce[:], signed)
	out, ok := secretbox.Open(nil, []byte(signed[24:]), &nonce, &b.key)
	if !ok {
		return "", errors.New("signature verification failed")
	}

	return string(out), nil
}

func (b *apibooter) readLocal(u *url.URL) (io.ReadCloser, error) {
	return os.Open(u.Path)
}

func (b *apibooter) readRemote(u string) (io.ReadCloser, error) {
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GET %q failed: %s", u, resp.Status)
	}
	return resp.Body, nil
}
