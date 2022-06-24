package server

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/bluemir/0xC0DE/internal/static"
)

func NewRenderer() (*template.Template, error) {
	tmpl := template.New("__root__")

	tmplFS, err := fs.Sub(static.Static, "html-templates")
	if err != nil {
		return nil, err
	}
	fs.WalkDir(tmplFS, "/", func(path string, info fs.DirEntry, err error) error {
		if info == nil {
			logrus.Tracef("%#v %s", info, path)
			return nil
		}
		if info.IsDir() && info.Name()[0] == '.' && path != "/" {
			return filepath.SkipDir
		}
		if info.IsDir() || info.Name()[0] == '.' || !strings.HasSuffix(path, ".html") {
			return nil
		}
		logrus.Debugf("parse template: path: %s", path)

		buf, err := fs.ReadFile(tmplFS, path)
		if err != nil {
			return err
		}

		tmpl, err = tmpl.Parse(string(buf))
		if err != nil {
			return err
		}
		return nil
	})

	return tmpl, nil
}
