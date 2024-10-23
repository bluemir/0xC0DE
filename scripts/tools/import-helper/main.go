package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {

	var source string
	var target string

	flag.StringVar(&source, "dir", "", "")
	flag.StringVar(&target, "target", "", "target-file")

	flag.Parse()

	selfRef, err := filepath.Rel(source, target)
	if err != nil {
		fmt.Println(err)
		return
	}

	jsFiles := []string{}
	if err := filepath.Walk(source, func(p string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(p, ".js") {
			return nil
		}
		path, err := filepath.Rel(source, p)
		if err != nil {
			return err
		}

		if path == selfRef { // remove self reference
			return nil
		}

		jsFiles = append(jsFiles, path)

		return nil
	}); err != nil {
		panic(err)
	}

	lines := []string{}

	for _, file := range jsFiles {
		lines = append(lines, fmt.Sprintf(`import "./%s";`, file))
	}

	data := strings.Join(lines, "\n")

	if err := os.WriteFile(target, []byte(data), 0644); err != nil {
		panic(err)
	}
}
