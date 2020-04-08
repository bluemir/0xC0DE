package resources

import (
	rice "github.com/GeertJohan/go.rice"
)

var (
	Static        = rice.MustFindBox("../../build/static")
	HTMLTemplates = rice.MustFindBox("../../resources/html-templates")
)
