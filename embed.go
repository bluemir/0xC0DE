//go:build !noembed

package main

import (
	"embed"

	"github.com/bluemir/0xC0DE/internal/static"
)

//go:embed build/static/*
var statics embed.FS

//go:embed build/templates/*
var templates embed.FS

func init() {
	static.InitFS(statics)
}
