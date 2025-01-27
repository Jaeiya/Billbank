package utils

import tea "github.com/charmbracelet/bubbletea"

type CommandModelMsg interface {
	Init() tea.Cmd
	Update(tea.Msg) (CommandModelMsg, tea.Cmd)
	View() string
	IsImplemented() bool
}

type ViewportSizeMsg struct {
	Width  int
	Height int
}
