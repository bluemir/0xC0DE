package server

import (
	"encoding/json"
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/bluemir/0xC0DE/internal/static"
)

func NewRenderer() (*template.Template, error) {
	tmpl := template.New("__root__").Funcs(template.FuncMap{
		"join": strings.Join,
		"json": json.Marshal,
		"toString": func(buf []byte) string {
			return string(buf)
		},
	})

	if err := fs.WalkDir(static.Templates, ".", func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return errors.Wrapf(err, "read template error: path: %s", path)
		}
		logrus.Debugf("read template: path: %s", path)

		if info.IsDir() && info.Name()[0] == '.' && path != "/" {
			return filepath.SkipDir
		}
		if info.IsDir() || info.Name()[0] == '.' || !strings.HasSuffix(path, ".html") {
			return nil
		}
		logrus.Debugf("parse template: path: %s", path)

		buf, err := fs.ReadFile(static.Templates, path)
		if err != nil {
			return err
		}

		tmpl, err = tmpl.New(path).Parse(string(buf))
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	for _, t := range tmpl.Templates() {
		logrus.Tracef("there is '%s' template", t.Name())
	}

	return tmpl, nil
}
