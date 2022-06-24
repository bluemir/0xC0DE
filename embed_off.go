//go:build noembed

package main

import (
	"os"

	"github.com/bluemir/0xC0DE/internal/static"
)

func init() {
	// default, when no embed.
	if err := static.InitFS(os.DirFS("./")); err != nil {
		panic(err)
	}
}
