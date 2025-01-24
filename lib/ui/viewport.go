package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jaeiya/billbank/lib/commands"
)

type CurrentCmd struct {
	Model tea.Model
	Cmd   tea.Cmd
}

type ViewPort struct {
	Commander       CmdInputModel
	CurrentCmdModel commands.CommandModelMsg
	height          int
	width           int
	status          string
}

func (vp ViewPort) Init() tea.Cmd {
	return nil
}

func (vp ViewPort) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case commands.CommandModelMsg:
		vp.CurrentCmdModel = msg
	case tea.WindowSizeMsg:
		vp.height = msg.Height
		vp.width = msg.Width
	}
	if vp.CurrentCmdModel != nil {
		vp.CurrentCmdModel, _ = vp.CurrentCmdModel.Update(msg)
	}
	vp.Commander, cmd = vp.Commander.Update(msg)
	return vp, cmd
}

func (vp ViewPort) View() string { // Define a style with a fixed height and bottom alignment
	cmdrStr := vp.Commander.View()
	h := lipgloss.Height(cmdrStr)
	cmdView := ""
	if vp.CurrentCmdModel != nil {
		cmdView = vp.CurrentCmdModel.View()
	}
	block := lipgloss.Place(vp.width, vp.height-h, lipgloss.Center, lipgloss.Center, cmdView)

	content := lipgloss.JoinVertical(lipgloss.Top, block, cmdrStr)

	return content
}
