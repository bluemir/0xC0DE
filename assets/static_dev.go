//go:build !prod

package assets

import (
	"fmt"
	"io/fs"
	"os"
)

// Static returns the filesystem for serving static files in dev mode.
// Path mapping: js/* -> src/js/*, css/* -> src/css/*, lib/* -> lib/*
func Static() fs.FS {
	return &mappedFS{fs: os.DirFS("assets"), prefix: src}
}

// CheckDevAssets verifies that required source files exist.
// This should be called at server startup to fail fast if files are missing.
func CheckDevAssets() error {
	requiredFiles := []string{
		"assets/src/js/index.js",
		"assets/src/css/page.css",
		"assets/lib/bm.js/bm.module.js",
	}
	for _, f := range requiredFiles {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			return fmt.Errorf("[DEV] Required asset file not found: %s\n"+
				"Make sure you're running from the project root directory", f)
		}
	}
	return nil
}
