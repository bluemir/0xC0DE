package server

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/bluemir/0xC0DE/pkg/static"
)

func NewRenderer() (*template.Template, error) {
	log := logrus.WithField("method", "NewRenderer")
	tmpl := template.New("__root__")

	static.HTMLTemplates.Walk("/", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && info.Name()[0] == '.' && path != "/" {
			return filepath.SkipDir
		}
		if info.IsDir() || info.Name()[0] == '.' || !strings.HasSuffix(path, ".html") {
			return nil
		}
		log.Debugf("parse template: path: %s", path)

		tmpl, err = tmpl.Parse(static.HTMLTemplates.MustString(path))
		if err != nil {
			return err
		}
		return nil
	})

	return tmpl, nil
}
