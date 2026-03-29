package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/bluemir/make-select/internal/parser"
)

type item struct {
	isCategory bool
	category   string
	target     parser.Target
}

type Model struct {
	items    []item
	cursor   int
	Selected string
}

func New(categories []parser.Category) Model {
	items := flatten(categories)
	return Model{
		items:  items,
		cursor: firstSelectableIndex(items),
	}
}

func flatten(categories []parser.Category) []item {
	var items []item
	for _, cat := range categories {
		if len(cat.Targets) == 0 {
			continue
		}
		items = append(items, item{isCategory: true, category: cat.Name})
		for _, t := range cat.Targets {
			items = append(items, item{target: t})
		}
	}
	return items
}

func firstSelectableIndex(items []item) int {
	for i, it := range items {
		if !it.isCategory {
			return i
		}
	}
	return 0
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "up", "k":
			m.cursor = m.prevSelectable()
		case "down", "j":
			m.cursor = m.nextSelectable()
		case "enter":
			if m.cursor < len(m.items) && !m.items[m.cursor].isCategory {
				m.Selected = m.items[m.cursor].target.Name
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m Model) prevSelectable() int {
	i := m.cursor - 1
	for i >= 0 {
		if !m.items[i].isCategory {
			return i
		}
		i--
	}
	return m.cursor
}

func (m Model) nextSelectable() int {
	i := m.cursor + 1
	for i < len(m.items) {
		if !m.items[i].isCategory {
			return i
		}
		i++
	}
	return m.cursor
}

var (
	categoryStyle = lipgloss.NewStyle().Bold(true).MarginTop(1)
	targetStyle   = lipgloss.NewStyle().PaddingLeft(2)
	nameStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Width(18)
	descStyle     = lipgloss.NewStyle()
	cursorStyle   = lipgloss.NewStyle().PaddingLeft(2).Bold(true)
	cursorName    = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Width(18).Bold(true)
)

func (m Model) View() string {
	var b strings.Builder

	b.WriteString("Select a make target:\n")

	for i, it := range m.items {
		if it.isCategory {
			name := it.category
			if name == "" {
				name = "General"
			}
			b.WriteString(categoryStyle.Render(name))
			b.WriteString("\n")
			continue
		}

		if i == m.cursor {
			b.WriteString(cursorStyle.Render(
				"▸ " + cursorName.Render(it.target.Name) + " " + descStyle.Render(it.target.Description),
			))
		} else {
			b.WriteString(targetStyle.Render(
				"  " + nameStyle.Render(it.target.Name) + " " + descStyle.Render(it.target.Description),
			))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n↑/↓: move  enter: select  q: quit\n")

	return b.String()
}
