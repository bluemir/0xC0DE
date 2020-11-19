package static

import (
	rice "github.com/GeertJohan/go.rice"
)

var (
	Static        = rice.MustFindBox("../../build/static")
	NodeModules   = rice.MustFindBox("../../node_modules")
	HTMLTemplates = rice.MustFindBox("../../static/html-templates")
)
