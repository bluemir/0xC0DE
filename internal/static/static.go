package static

import (
	"io/fs"
	"os"
)

var (
	Static fs.FS = os.DirFS("build/static") // default, when no embed.
)

func InitFS(rootfs fs.FS) {
	Static = rootfs
}
