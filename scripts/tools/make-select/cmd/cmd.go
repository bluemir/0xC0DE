package cmd

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

	"github.com/bluemir/make-select/internal/parser"
	"github.com/bluemir/make-select/internal/tui"
)

func Run() error {
	printOnly := flag.Bool("print-only", false, "print targets and exit")
	flag.Parse()

	root := findProjectRoot()

	categories := parser.Parse(root)
	if len(categories) == 0 {
		return fmt.Errorf("no targets found")
	}

	if *printOnly {
		printTargets(categories)
		return nil
	}

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		runMake(root, "build")
		return nil
	}

	m := tui.New(categories)

	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	final := result.(tui.Model)
	if final.Selected != "" {
		runMake(root, final.Selected)
	}
	return nil
}

func printTargets(categories []parser.Category) {
	fmt.Println("# Usage:")
	fmt.Println("#   make \033[36m<target>\033[0m")
	for _, cat := range categories {
		if len(cat.Targets) == 0 {
			continue
		}
		name := cat.Name
		if name == "" {
			name = "General"
		}
		fmt.Printf("#\n# \033[1m%s\033[0m\n", name)
		for _, t := range cat.Targets {
			fmt.Printf("#   \033[36m%-15s\033[0m %s\n", t.Name, t.Description)
		}
	}
	fmt.Println("#")
	fmt.Println("# This project used https://github.com/bluemir/0xC0DE as template.")
}

// findProjectRoot walks up from cwd looking for a Makefile
func findProjectRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "Makefile")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			wd, _ := os.Getwd()
			return wd
		}
		dir = parent
	}
}

func runMake(root, target string) {
	makeBin, err := exec.LookPath("make")
	if err != nil {
		fmt.Fprintln(os.Stderr, "make not found")
		os.Exit(1)
	}
	os.Chdir(root)
	syscall.Exec(makeBin, []string{"make", target}, os.Environ())
}
