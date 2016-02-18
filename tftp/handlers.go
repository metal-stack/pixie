package tftp

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
)

// FilesystemHandler returns a Handler that serves files in root.
func FilesystemHandler(root string) (Handler, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	root = filepath.ToSlash(root)
	return func(path string, addr net.Addr) (io.ReadCloser, error) {
		// Join with a root, which gets rid of directory traversal
		// attempts. Then we join that canonicalized path with the
		// actual root, which resolves to the actual on-disk file to
		// serve.
		path = filepath.Join("/", path)
		path = filepath.FromSlash(filepath.Join(root, path))

		st, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if !st.Mode().IsRegular() {
			return nil, fmt.Errorf("requested path %q is not a file", path)
		}
		return os.Open(path)
	}, nil
}

// ConstantHandler returns a Handler that serves bs for all requested paths.
func ConstantHandler(bs []byte) Handler {
	return func(path string, addr net.Addr) (io.ReadCloser, error) {
		return ioutil.NopCloser(bytes.NewBuffer(bs)), nil
	}
}
