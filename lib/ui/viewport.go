package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CurrentCmd struct {
	Model tea.Model
	Cmd   tea.Cmd
}

type ViewPort struct {
	Commander  CmdInputModel
	CurrentCmd *CurrentCmd
	height     int
	status     string
}

func (vp ViewPort) Init() tea.Cmd {
	return nil
}

func (vp ViewPort) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case TestMsg:
		vp.status = "I executed because of a command!!"
	case tea.WindowSizeMsg:
		vp.height = msg.Height
	}
	vp.Commander, cmd = vp.Commander.Update(msg)
	return vp, cmd
}

func (vp ViewPort) View() string {
	if vp.status != "" {
		return vp.status
	}

	cmdrStr := vp.Commander.View()
	h := lipgloss.Height(cmdrStr)
	padding := ""
	if vp.height > 0 {
		padding = strings.Repeat("\n", vp.height-h)
	}
	return padding + cmdrStr
}
