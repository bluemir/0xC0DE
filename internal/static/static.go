package static

import (
	"io/fs"
	"os"
)

var (
	Static    fs.FS = os.DirFS("build/static")    // default, when no embed.
	Templates fs.FS = os.DirFS("build/templates") // default, when no embed.
)

func InitFS(rootfs fs.FS) error {
	var err error

	Static, err = fs.Sub(rootfs, "build/static")
	if err != nil {
		return err
	}
	Templates, err = fs.Sub(rootfs, "build/templates")
	if err != nil {
		return err
	}
	return nil
}
