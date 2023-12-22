//go:build embed

package main

import (
	"embed"

	"github.com/bluemir/0xC0DE/internal/static"
)

//go:embed build/static
//go:embed assets/html-templates
var embedFS embed.FS

func init() {
	if err := static.InitFS(embedFS); err != nil {
		panic(err)
	}
}
