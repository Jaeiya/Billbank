package lib

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jaeiya/billbank/lib/cmd"
	"github.com/jaeiya/billbank/lib/logger"
	"github.com/jaeiya/billbank/lib/ui"
)

type (
	ActiveCmdMsg       string
	CommanderStatusMsg struct {
		String   string
		Severity cmd.StatusSeverity
	}
)

type ViewportSizeMsg struct {
	Width  int
	Height int
}

type ViewPort struct {
	CommandInput    cmd.CmdInputModel
	CurrentCmdModel cmd.Model
	CommandStatus   cmd.Status
	height          int
	width           int
}

func (vp ViewPort) Init() tea.Cmd {
	return nil
}

func (vp ViewPort) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var teaCmds []tea.Cmd
	var teaCmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		vp.height = msg.Height
		vp.width = msg.Width
		logger.Log(logger.Hot, "ViewPort", "[WindowSizeMsg] sending viewport size [%d:%d]", vp.width, vp.height)
		teaCmds = append(teaCmds, vp.sendViewportSize)

	case cmd.UpdateCmdMsg:
		if msg.Model != nil {
			vp.CurrentCmdModel = msg.Model
			logger.Log(logger.Debug, "ViewPort", "storing new command model [%s]", msg.Status.BranchStr)
		}
		logger.Log(logger.Debug, "ViewPort", "[UpdateCmdMsg] sending viewport size [%d:%d]", vp.width, vp.height)
		teaCmds = append(teaCmds, vp.sendViewportSize)

	case CommanderStatusMsg:
		teaCmds = append(teaCmds, vp.sendStatusMsg(msg.String, msg.Severity))
	}

	vp.CommandInput, teaCmd = vp.CommandInput.Update(msg)
	teaCmds = append(teaCmds, teaCmd)

	if vp.CurrentCmdModel != nil {
		vp.CurrentCmdModel, teaCmd = vp.CurrentCmdModel.Update(msg)
		teaCmds = append(teaCmds, teaCmd)
	}

	return vp, tea.Batch(teaCmds...)
}

func (vp ViewPort) View() string {
	cmdrStr := vp.CommandInput.View()
	h := lipgloss.Height(cmdrStr)
	cmdView := ""

	if vp.CurrentCmdModel != nil && vp.CurrentCmdModel.IsInitialized() {
		cmdView = vp.CurrentCmdModel.View()
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

func (vp ViewPort) sendViewportSize() tea.Msg {
	return cmd.ViewportSizeMsg{
		Height: vp.height - lipgloss.Height(vp.CommandInput.View()),
		Width:  vp.width,
	}
}

func (vp ViewPort) sendStatusMsg(msg string, s cmd.StatusSeverity) func() tea.Msg {
	return func() tea.Msg {
		return cmd.UpdateStatusMsg{
			String:   msg,
			Severity: s,
		}
	}
}
