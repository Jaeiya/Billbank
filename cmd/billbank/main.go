package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct{}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	return "hello world"
}

func main() {
	if _, err := tea.NewProgram(Model{}).Run(); err != nil {
		fmt.Println("Error running program")
	}
}
