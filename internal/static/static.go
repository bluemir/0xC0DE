package static

import (
	rice "github.com/GeertJohan/go.rice"
)

var (
	Static        = rice.MustFindBox("../../build/static")
	HTMLTemplates = rice.MustFindBox("../../web/html-templates")
	// NodeModules   = rice.MustFindBox("../../node_modules")
)
