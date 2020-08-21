package static

import (
	rice "github.com/GeertJohan/go.rice"
)

var (
	Static        = rice.MustFindBox("../../build/static")
	HTMLTemplates = rice.MustFindBox("../../static/html-templates")
)
