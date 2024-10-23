package assets

import (
	"embed"
)

//go:embed js/* css/* lib/*
var Static embed.FS

//go:embed html-templates
var HtmlTemplates embed.FS
