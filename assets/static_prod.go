//go:build prod

package assets

import (
	"embed"
	"io/fs"
)

// go generate로 esbuild 실행 (prod 빌드 시)
//go:generate mkdir -p dist/js dist/css
//go:generate go tool esbuild src/js/index.js --outdir=dist/js --bundle --minify --format=esm --external:lit-html --external:bm.js/bm.module.js --alias:@=./src/js
//go:generate go tool esbuild src/css/page.css src/css/element.css --outdir=dist/css --bundle --minify

//go:embed dist/js/* dist/css/* bundle/*
var staticFS embed.FS

// Static returns the embedded filesystem with path mapping:
// dist/js/* -> js/*, dist/css/* -> css/*
func Static() fs.FS {
	return &mappedFS{fs: staticFS, prefix: dist}
}

// CheckDevAssets is a no-op in prod mode since files are embedded
func CheckDevAssets() error {
	return nil
}

//go:embed html-templates
var HtmlTemplates embed.FS
