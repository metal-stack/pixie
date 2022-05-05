package ipxe

import (
	"embed"
	_ "embed"
	"path"
)

//go:embed bin
var payload embed.FS

func MustGet(name string) []byte {
	contents, err := payload.ReadFile(path.Join("bin", name))
	if err != nil {
		panic(err)
	}
	return contents
}
