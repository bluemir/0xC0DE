//go:build !noembed

package main

import (
	"embed"
	"io/fs"

	"github.com/bluemir/0xC0DE/internal/static"
)

//go:embed build/static/*
//go:embed build/templates/*
var embedFS embed.FS

func init() {
	statics, err := fs.Sub(embedFS, "build/static")
	if err != nil {
		panic(err)
	}
	templates, err := fs.Sub(embedFS, "build/templates")
	if err != nil {
		panic(err)
	}
	static.InitFS(statics, templates)
}
