package utils

import tea "github.com/charmbracelet/bubbletea"

type CommandModelMsg struct {
	ID    int
	Model CommandModel
}

type CommandModel interface {
	Init() tea.Cmd
	Update(tea.Msg) (CommandModel, tea.Cmd)
	View() string
	IsImplemented() bool
}

type CommandStrMsg string

type ViewportSizeMsg struct {
	Width  int
	Height int
}
