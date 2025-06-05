package ipxe

import (
	"embed"
	"path"
)

// Files to embed get created during make ipxe
// nolint:typecheck
//
//go:embed ipxe/bin
var payload embed.FS

func MustGet(name string) []byte {
	contents, err := payload.ReadFile(path.Join("ipxe/bin", name))
	if err != nil {
		panic(err)
	}
	return contents
}
