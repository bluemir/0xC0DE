package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/sirupsen/logrus"
)

type final struct{}

func (m final) Init() tea.Cmd                           { return tea.Quit }
func (m final) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, tea.Quit }
func (m final) View() tea.View                          { return tea.NewView("") }

type finalAction interface {
	Action(ctx context.Context) error
}

// --- exit with error ---

func ExitWithError(err error) (tea.Model, tea.Cmd) {
	return finalError{err: err}, tea.Quit
}

type finalError struct {
	final

	err error
}

func (m finalError) Action(ctx context.Context) error {
	return m.err
}

// --- exit without error ---

func Quit() (tea.Model, tea.Cmd) {
	return finalQuit{}, tea.Quit
}
func Exit() (tea.Model, tea.Cmd) {
	return finalQuit{}, tea.Quit
}

type finalQuit struct {
	final
}

func (m finalQuit) Action(ctx context.Context) error {
	logrus.Info("bye")
	return nil
}
