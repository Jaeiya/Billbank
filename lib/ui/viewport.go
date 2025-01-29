package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jaeiya/billbank/lib/utils"
)

type ActiveCmdMsg string

type ViewportSizeMsg struct {
	Width  int
	Height int
}

type CurrentCmd struct {
	Model tea.Model
	Cmd   tea.Cmd
}

type ViewPort struct {
	Commander       CmdInputModel
	CurrentCmdModel CommandModel
	CommandStatus   CommandStatus
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
	case CommandMsg:
		vp.CurrentCmdModel = msg
		vp.CurrentCmdModel, _ = vp.CurrentCmdModel.Update(vp.sendViewportSize())
		vp.CurrentCmdModel, _ = vp.CurrentCmdModel.Update(vp.sendActiveCmd(msg.status.CommandStr))

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
	return ViewportSizeMsg{
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

func (vp ViewPort) sendActiveCmd(cmd string) tea.Msg {
	return ActiveCmdMsg(cmd)
}
