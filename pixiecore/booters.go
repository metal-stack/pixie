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
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"text/template"
	"time"

	v1 "github.com/metal-stack/metal-api/pkg/api/v1"
	"github.com/metal-stack/pixie/api"
)

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
		return nil, fmt.Errorf("failed to get randomness for signing key: %w", err)
	}

	return ret, nil
}
func GRPCBooter(log *slog.Logger, client *GrpcClient, partition string, metalAPIConfig *api.MetalConfig) (Booter, error) {
	ret := &grpcbooter{
		grpc:      client,
		partition: partition,
		log:       log,
		config:    metalAPIConfig,
	}
	if _, err := io.ReadFull(rand.Reader, ret.key[:]); err != nil {
		return nil, fmt.Errorf("failed to get randomness for signing key: %w", err)
	}
	log.Info("starting grpc booter", "partition", partition)
	return ret, nil
}

type apibooter struct {
	client    *http.Client
	urlPrefix string
	key       [32]byte
}

type grpcbooter struct {
	apibooter
	grpc      *GrpcClient
	config    *api.MetalConfig
	partition string
	log       *slog.Logger
}

// BootSpec implements Booter
func (g *grpcbooter) BootSpec(m Machine) (*Spec, error) {
	g.log.Info("bootspec", "machine", m.String())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var r rawSpec
	if m.GUID != "" {
		// Very first dhcp call which contains Machine UUID, tell metal-api this uuid
		req := &v1.BootServiceDhcpRequest{
			Uuid: string(m.GUID),
		}
		g.log.Info("dhcp", "req", req)
		_, err := g.grpc.BootService().Dhcp(ctx, req)
		if err != nil {
			g.log.Error("boot", "error", err)
			return nil, err
		}
		r = rawSpec{}
	} else {
		// machine asks for a dhcp answer, ask metal-api for a proper response in this partition
		req := &v1.BootServiceBootRequest{
			Mac:         m.MAC.String(),
			PartitionId: g.partition,
		}
		g.log.Info("boot", "req", req)
		resp, err := g.grpc.BootService().Boot(ctx, req)
		if err != nil {
			g.log.Error("boot", "error", err)
			return nil, err
		}
		g.log.Info("boot", "resp", resp)

		cmdline := []string{resp.GetCmdline(), fmt.Sprintf("PIXIE_API_URL=%s", g.config.PixieAPIURL)}
		if g.config.Debug {
			cmdline = append(cmdline, "DEBUG=1")
		}

		r = rawSpec{
			Kernel:  resp.GetKernel(),
			Initrd:  resp.GetInitRamDisks(),
			Cmdline: strings.Join(cmdline, " "),
		}
	}

	spec, err := bootSpec(g.key, g.urlPrefix, r)
	g.log.Info("bootspec", "raw spec", r, "return spec", spec)
	return spec, err
}

func (b *apibooter) getAPIResponse(m Machine) (io.ReadCloser, error) {
	var reqURL string
	reqURL = fmt.Sprintf("%s/boot/%s", b.urlPrefix, m.MAC)
	if m.GUID != "" {
		reqURL = fmt.Sprintf("%s/dhcp/%s", b.urlPrefix, m.GUID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("%s: %s", reqURL, http.StatusText(resp.StatusCode))
	}

	return resp.Body, nil
}

func (b *apibooter) BootSpec(m Machine) (*Spec, error) {
	body, err := b.getAPIResponse(m)
	if body != nil {
		defer func() {
			_ = body.Close()
		}()
	}
	if err != nil {
		return nil, err
	}

	var r rawSpec
	if err = json.NewDecoder(body).Decode(&r); err != nil {
		return nil, err
	}

	return bootSpec(b.key, b.urlPrefix, r)
}

type rawSpec struct {
	Kernel     string   `json:"kernel"`
	Initrd     []string `json:"initrd"`
	Cmdline    any      `json:"cmdline"`
	Message    string   `json:"message"`
	IpxeScript string   `json:"ipxe-script"`
}

func bootSpec(key [32]byte, prefix string, r rawSpec) (*Spec, error) {
	if r.IpxeScript != "" {
		return &Spec{
			IpxeScript: r.IpxeScript,
		}, nil
	}

	var err error
	r.Kernel, err = makeURLAbsolute(prefix, r.Kernel)
	if err != nil {
		return nil, err
	}
	for i, img := range r.Initrd {
		r.Initrd[i], err = makeURLAbsolute(prefix, img)
		if err != nil {
			return nil, err
		}
	}

	ret := Spec{
		Message: r.Message,
	}
	if ret.Kernel, err = signURL(r.Kernel, &key); err != nil {
		return nil, err
	}
	for _, img := range r.Initrd {
		initrd, err := signURL(img, &key)
		if err != nil {
			return nil, err
		}
		ret.Initrd = append(ret.Initrd, initrd)
	}

	if r.Cmdline != nil {
		switch c := r.Cmdline.(type) {
		case string:
			ret.Cmdline = c
		case map[string]any:
			ret.Cmdline, err = constructCmdline(c)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("API server returned unknown type %T for kernel cmdline", r.Cmdline)
		}
	}

	f := func(u string) (string, error) {
		urlStr, err := makeURLAbsolute(prefix, u)
		if err != nil {
			return "", fmt.Errorf("invalid url %q for cmdline: %w", urlStr, err)
		}
		id, err := signURL(urlStr, &key)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("{{ ID %q }}", id), nil
	}
	ret.Cmdline, err = expandCmdline(ret.Cmdline, template.FuncMap{"URL": f})
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (b *apibooter) ReadBootFile(id ID) (io.ReadCloser, int64, error) {
	urlStr, err := getURL(id, &b.key)
	if err != nil {
		return nil, -1, err
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, -1, fmt.Errorf("%q is not an URL", urlStr)
	}
	var (
		ret io.ReadCloser
		sz  int64
	)
	if u.Scheme == "file" {
		// TODO serveFile
		f, err := os.Open(u.Path)
		if err != nil {
			return nil, -1, err
		}
		fi, err := f.Stat()
		if err != nil {
			_ = f.Close()
			return nil, -1, err
		}
		ret, sz = f, fi.Size()
	} else {
		// urlStr will get reparsed by http.Get, which is mildly
		// wasteful, but the code looks nicer than constructing a
		// Request.
		resp, err := http.Get(urlStr) // nolint:gosec,bodyclose,noctx
		if err != nil {
			return nil, -1, err
		}
		if resp.StatusCode != 200 {
			return nil, -1, fmt.Errorf("GET %q failed: %s", urlStr, resp.Status)
		}

		ret = resp.Body
		sz = resp.ContentLength
	}
	return ret, sz, nil
}

func (b *apibooter) WriteBootFile(id ID, body io.Reader) error {
	u, err := getURL(id, &b.key)
	if err != nil {
		return err
	}

	resp, err := http.Post(u, "application/octet-stream", body) // nolint:gosec,noctx
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("POST %q failed: %s", u, resp.Status)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	return nil
}

func makeURLAbsolute(prefix, urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("%q is not an URL", urlStr)
	}
	if !u.IsAbs() {
		base, err := url.Parse(prefix)
		if err != nil {
			return "", err
		}
		u = base.ResolveReference(u)
	}
	return u.String(), nil
}

func constructCmdline(m map[string]any) (string, error) {
	var c []string
	for k := range m {
		c = append(c, k)
	}
	sort.Strings(c)

	var ret []string
	for _, k := range c {
		switch v := m[k].(type) {
		case bool:
			ret = append(ret, k)
		case string:
			ret = append(ret, fmt.Sprintf("%s=%q", k, v))
		case map[string]any:
			urlStr, ok := v["url"].(string)
			if !ok {
				return "", fmt.Errorf("cmdline key %q has object value with no 'url' attribute", k)
			}
			ret = append(ret, fmt.Sprintf("%s={{ URL %q }}", k, urlStr))
		default:
			return "", fmt.Errorf("unsupported value kind %T for cmdline key %q", m[k], k)
		}
	}
	return strings.Join(ret, " "), nil
}
