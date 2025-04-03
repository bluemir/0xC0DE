package server

import (
	"encoding/json"
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/bluemir/0xC0DE/assets"
	"github.com/bluemir/0xC0DE/internal/server/middleware/cache"
)

func NewRenderer() (*template.Template, error) {
	tmpl := template.New("__root__").Funcs(template.FuncMap{
		"join": strings.Join,
		"json": json.Marshal,
		"toString": func(buf []byte) string {
			return string(buf)
		},
		"rev": func(c *gin.Context) string {
			return c.GetString(cache.REVVED)
		},
	})

	templates, err := fs.Sub(assets.HtmlTemplates, "html-templates")
	if err != nil {
		return nil, err
	}

	if err := fs.WalkDir(templates, ".", func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return errors.Wrapf(err, "read template error: path: %s", path)
		}
		logrus.Debugf("read template: path: %s", path)

		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && path != "." {
			return filepath.SkipDir
		}

		if info.IsDir() || strings.HasPrefix(info.Name(), ".") || !strings.HasSuffix(path, ".html") {
			return nil
		}
		logrus.Debugf("parse template: path: %s", path)

		buf, err := fs.ReadFile(templates, path)
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
