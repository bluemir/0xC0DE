package static

import (
	"io/fs"
	"os"
)

var (
	Static    fs.FS = os.DirFS("build/static")    // default, when no embed.
	Templates fs.FS = os.DirFS("build/templates") // default, when no embed.
)

func InitFS(rootfs fs.FS, templates fs.FS) {
	Static = rootfs
	Templates = templates
}
