package lib

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jaeiya/billbank/lib/commander"
	"github.com/jaeiya/billbank/lib/ui"
	"github.com/jaeiya/billbank/lib/utils/logger"
)

type (
	ActiveCmdMsg       string
	CommanderStatusMsg struct {
		String   string
		Severity commander.StatusSeverity
	}
)

type ViewportSizeMsg struct {
	Width  int
	Height int
}

type ViewPort struct {
	Commander       commander.CmdInputModel
	CurrentCmdModel commander.CommandModel
	CommandStatus   commander.CommandStatus
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
	case tea.WindowSizeMsg:
		vp.height = msg.Height
		vp.width = msg.Width

	case commander.CommandModel:
		vp.CurrentCmdModel = msg
		cmds = append(cmds, vp.sendViewportSize)

	case CommanderStatusMsg:
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

	if vp.CurrentCmdModel != nil {
		cmdView = vp.CurrentCmdModel.View()
	}

	if vp.lastCmdError != nil {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			ui.NewErrorBox("Command Error", vp.lastCmdError.Error(), vp.width, vp.height-h),
			cmdrStr,
		)
	}

	return lipgloss.JoinVertical(
		lipgloss.Top,
		lipgloss.NewStyle().Foreground(ui.FgColor).Render(
			lipgloss.Place(
				vp.width,
				vp.height-h,
				lipgloss.Left,
				lipgloss.Top,
				cmdView,
			),
		),
		cmdrStr,
	)
}

func (vp *ViewPort) catchCmdErrors() tea.Cmd {
	err := vp.CurrentCmdModel.GetError()

	if err == vp.lastCmdError {
		return nil
	}

	if err != nil && vp.lastCmdError != nil {
		if err.Error() == vp.lastCmdError.Error() {
			return nil
		}
	}

	if err != nil {
		logger.Log(logger.Error, fmt.Sprintf("CommandError: %s", err))
		vp.lastCmdError = err
		return vp.sendStatusMsg("Command Implementation Error", commander.HIGH)
	}

	vp.lastCmdError = nil

	return nil
}

func (vp ViewPort) sendViewportSize() tea.Msg {
	return commander.CmdViewportSizeMsg{
		Height: vp.height - lipgloss.Height(vp.Commander.View()),
		Width:  vp.width,
	}
}

func (vp ViewPort) sendStatusMsg(msg string, s commander.StatusSeverity) func() tea.Msg {
	return func() tea.Msg {
		return commander.UpdateStatusMsg{
			String:   msg,
			Severity: s,
		}
	}
}
