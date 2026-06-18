package tui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/cockroachdb/errors"
)

func Run(ctx context.Context) error {

	final, err := tea.NewProgram(
		&viewMainMenu{},
		tea.WithContext(ctx),
	).Run()

	if err != nil {
		return err
	}

	action, ok := final.(finalAction)
	if !ok {
		return errors.Newf("intro ended on unexpected model %T", final)
	}
	switch err := action.Action(ctx); {
	//case errors.Is(err, ...):
	//	return nil
	//case errors.Is(err, ...):
	// return nil
	case err != nil:
		return err
	}
	return nil
}

type viewMainMenu struct {
	cursor int
}

func (m viewMainMenu) Init() tea.Cmd {
	return nil
}
func (m viewMainMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return QuitConfirm(m)
		case "q":
			m.cursor = 1
			return m, tea.Quit
		case "enter":
			switch m.cursor {
			case 0:
				fmt.Printf("world")
				return m, nil
			case 1:
				return QuitConfirm(m)
			default:
				return ExitWithError(errors.Errorf("Invalid state"))
			}
		default:
			return m, nil
		}
	default:
		return m, nil
	}
}
func (m viewMainMenu) View() tea.View {
	return tea.NewView(lipgloss.JoinVertical(
		lipgloss.Left,
		cursor(m.cursor == 0, "hello"),
		cursor(m.cursor == 1, "exit"),
	))
}

func QuitConfirm(parent tea.Model) (tea.Model, tea.Cmd) {
	return viewQuitConfirm{parent: parent}, nil
}

type viewQuitConfirm struct {
	parent tea.Model
	cursor int
}

func (m viewQuitConfirm) Init() tea.Cmd {
	return nil
}

func (m viewQuitConfirm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return Exit()
		case "right", "y":
			m.cursor = 0
			return m, nil
		case "left", "n":
			m.cursor = 1
			return m, nil
		case "esc":
			return m.parent, nil
		case "enter":
			switch m.cursor {
			case 0:
				return Quit()
			case 1:
				return m.parent, nil
			default:
				return ExitWithError(errors.Errorf("Invalid state"))
			}
		default:
			return m, nil
		}
	default:
		return m, nil
	}
}
func (m viewQuitConfirm) View() tea.View {
	style := lipgloss.NewStyle().Padding(2)
	return tea.NewView(style.Render(lipgloss.JoinHorizontal(
		lipgloss.Top,
		cursor(m.cursor == 0, "Yes"),
		cursor(m.cursor == 1, "No"),
	)))
}
