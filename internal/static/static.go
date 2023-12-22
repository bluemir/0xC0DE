package static

import (
	"io/fs"
)

var (
	Static    fs.FS
	Templates fs.FS
)

func InitFS(rootfs fs.FS) error {
	var err error

	Static, err = fs.Sub(rootfs, "build/static")
	if err != nil {
		return err
	}
	Templates, err = fs.Sub(rootfs, "assets/html-templates")
	if err != nil {
		return err
	}
	return nil
}
