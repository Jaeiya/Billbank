package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	commands "github.com/jaeiya/billbank/lib/ui/commands"
	"github.com/jaeiya/billbank/lib/utils"
)

type CurrentCmd struct {
	Model tea.Model
	Cmd   tea.Cmd
}

type ViewPort struct {
	Commander       CmdInputModel
	CurrentCmdModel commands.CommandModel
	height          int
	width           int
}

func (vp ViewPort) Init() tea.Cmd {
	return nil
}

func (vp ViewPort) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case commands.CommandModelMsg:
		if !msg.Model.IsImplemented() {
			utils.Log(utils.Error, "CommandError: command not implemented")
			return vp, vp.sendStatusMsg("Command Not Implemented", HIGH)
		}
		vp.CurrentCmdModel = msg.Model
		// Immediately send viewport size
		vp.CurrentCmdModel, _ = vp.CurrentCmdModel.Update(vp.sendViewportSize())

	case tea.WindowSizeMsg:
		vp.height = msg.Height
		vp.width = msg.Width
		cmds = append(cmds, vp.sendViewportSize)

	}

	if vp.CurrentCmdModel != nil {
		vp.CurrentCmdModel, cmd = vp.CurrentCmdModel.Update(msg)
		cmds = append(cmds, cmd)
	}

	vp.Commander, cmd = vp.Commander.Update(msg)
	cmds = append(cmds, cmd)

	return vp, tea.Batch(cmds...)
}

func (vp ViewPort) View() string {
	cmdrStr := vp.Commander.View()
	h := lipgloss.Height(cmdrStr)
	cmdView := ""
	if vp.CurrentCmdModel != nil {
		cmdView = vp.CurrentCmdModel.View()
	}

	block := lipgloss.Place(vp.width, vp.height-h, lipgloss.Left, lipgloss.Top, cmdView)
	content := lipgloss.JoinVertical(lipgloss.Top, block, cmdrStr)

	return content
}

func (vp ViewPort) sendViewportSize() tea.Msg {
	return utils.ViewportSizeMsg{
		Height: vp.height - lipgloss.Height(vp.Commander.View()),
		Width:  vp.width,
	}
}

func (vp ViewPort) sendStatusMsg(msg string, s StatusSeverity) func() tea.Msg {
	return func() tea.Msg {
		return UpdateStatusMsg{
			msg,
			s,
		}
	}
}
