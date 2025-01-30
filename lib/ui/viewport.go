package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jaeiya/billbank/lib/utils"
)

type (
	ActiveCmdMsg string
	CmdStatusMsg struct {
		String   string
		Severity StatusSeverity
	}
)

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
	lastCmdError    error
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
		utils.Log(utils.Info, "ViewPort: setting up command")
		vp.CurrentCmdModel = msg
		vp.CurrentCmdModel, _ = vp.CurrentCmdModel.Update(vp.sendViewportSize())
		vp.CurrentCmdModel, _ = vp.CurrentCmdModel.Update(vp.sendActiveCmd(msg.status.CommandStr))

	case tea.WindowSizeMsg:
		vp.height = msg.Height
		vp.width = msg.Width
		cmds = append(cmds, vp.sendViewportSize)

	case CmdStatusMsg:
		cmds = append(cmds, vp.sendStatusMsg(msg.String, msg.Severity))

	}

	if vp.CurrentCmdModel != nil {
		vp.CurrentCmdModel, cmd = vp.CurrentCmdModel.Update(msg)
		cmds = append(cmds, cmd, vp.catchCmdErrors())
	}

	vp.Commander, cmd = vp.Commander.Update(msg)
	cmds = append(cmds, cmd)

	return vp, tea.Batch(cmds...)
}

func (vp ViewPort) View() string {
	cmdrStr := vp.Commander.View()
	h := lipgloss.Height(cmdrStr)
	cmdView := ""
	alignX := lipgloss.Left
	alignY := lipgloss.Top

	if vp.CurrentCmdModel != nil {
		cmdView = vp.CurrentCmdModel.View()
	}

	if vp.lastCmdError != nil {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			NewErrorMsg("Command Error", vp.lastCmdError.Error(), vp.width, vp.height-h),
			cmdrStr,
		)
	}

	block := lipgloss.Place(
		vp.width,
		vp.height-h,
		alignX,
		alignY,
		cmdView,
	)
	content := lipgloss.JoinVertical(
		lipgloss.Top,
		lipgloss.NewStyle().Foreground(lipgloss.Color("#EEE")).Render(block),
		cmdrStr,
	)

	return content
}

func (vp *ViewPort) catchCmdErrors() tea.Cmd {
	err := vp.CurrentCmdModel.GetLastError()

	if err == vp.lastCmdError {
		return nil
	}

	if err != nil && vp.lastCmdError != nil {
		if err.Error() == vp.lastCmdError.Error() {
			return nil
		}
	}

	if err != nil {
		utils.Log(utils.Info, fmt.Sprintf("CommandError: %s", err))
		vp.lastCmdError = err
		return vp.sendStatusMsg("Command Implementation Error", HIGH)
	}

	vp.lastCmdError = nil

	return nil
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
