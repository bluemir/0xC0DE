//go:build !prod

package assets

import (
	"io/fs"
	"os"
)

var HtmlTemplates fs.FS = os.DirFS("assets")
