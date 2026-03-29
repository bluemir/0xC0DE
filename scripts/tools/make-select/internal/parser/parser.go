package parser

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Target struct {
	Name        string
	Description string
}

type Category struct {
	Name    string
	Targets []Target
}

var (
	targetRe   = regexp.MustCompile(`^([a-zA-Z_0-9-]+):.*?##\s*(.*)`)
	categoryRe = regexp.MustCompile(`^##@\s*(.*)`)
	includeRe  = regexp.MustCompile(`^-?include\s+(.+)`)
)

func Parse(root string) []Category {
	mainFile := filepath.Join(root, "Makefile")

	// collect files to parse: Makefile first, then included files
	files := []string{mainFile}

	// scan Makefile for include directives
	if f, err := os.Open(mainFile); err == nil {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if m := includeRe.FindStringSubmatch(line); m != nil {
				pattern := strings.TrimSpace(m[1])
				// expand make variables we know about
				pattern = strings.ReplaceAll(pattern, "$(wildcard ", "")
				pattern = strings.TrimSuffix(pattern, ")")
				pattern = strings.TrimSpace(pattern)

				absPattern := pattern
				if !filepath.IsAbs(pattern) {
					absPattern = filepath.Join(root, pattern)
				}
				if matches, err := filepath.Glob(absPattern); err == nil {
					files = append(files, matches...)
				}
			}
		}
		f.Close()
	}

	var categories []Category
	catIndex := map[string]int{}
	currentCat := ""

	for _, file := range files {
		parseFile(file, &currentCat, &categories, catIndex)
	}

	return categories
}

func parseFile(path string, currentCat *string, categories *[]Category, catIndex map[string]int) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		if m := categoryRe.FindStringSubmatch(line); m != nil {
			*currentCat = strings.TrimSpace(m[1])
			continue
		}

		if m := targetRe.FindStringSubmatch(line); m != nil {
			name := m[1]
			desc := strings.TrimSpace(m[2])

			cat := *currentCat
			idx, exists := catIndex[cat]
			if !exists {
				idx = len(*categories)
				catIndex[cat] = idx
				*categories = append(*categories, Category{Name: cat})
			}
			(*categories)[idx].Targets = append((*categories)[idx].Targets, Target{Name: name, Description: desc})
		}
	}
}
