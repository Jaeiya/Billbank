package lib

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jaeiya/billbank/lib/cmdmodel"
	"github.com/jaeiya/billbank/lib/logger"
	"github.com/jaeiya/billbank/lib/ui"
)

type (
	ActiveCmdMsg       string
	CommanderStatusMsg struct {
		String   string
		Severity cmdmodel.StatusSeverity
	}
)

type ViewportSizeMsg struct {
	Width  int
	Height int
}

type ViewPort struct {
	CommandInput    cmdmodel.InputModel
	CurrentCmdModel cmdmodel.Interface
	CommandStatus   cmdmodel.Status
	height          int
	width           int
}

func (vp ViewPort) Init() tea.Cmd {
	teaCmds := []tea.Cmd{
		textinput.Blink,
		vp.CommandInput.Init(),
	}
	return tea.Batch(teaCmds...)
}

func (vp ViewPort) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var teaCmds []tea.Cmd
	var teaCmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		vp.height = msg.Height
		vp.width = msg.Width
		logger.Log(logger.Hot, "[WindowSizeMsg] sending viewport size [%d:%d]", vp.width, vp.height)
		teaCmds = append(teaCmds, vp.sendViewportSize)

	case cmdmodel.ReleaseInputMsg:
		vp.CommandStatus.CaptureInput = false
		teaCmds = append(teaCmds, func() tea.Msg { return textinput.Blink() })

	case cmdmodel.UpdateCmdMsg:
		if msg.Model != nil {
			vp.CurrentCmdModel = msg.Model
			logger.Log(logger.Debug, "storing new command model [%s]", msg.CommandStatus.Path)
		}
		vp.CommandStatus = msg.CommandStatus
		logger.Log(logger.Debug, "[UpdateCmdMsg] sending viewport size [%d:%d]", vp.width, vp.height)
		teaCmds = append(teaCmds, vp.sendViewportSize)

	case CommanderStatusMsg:
		teaCmds = append(teaCmds, vp.sendStatusMsg(msg.String, msg.Severity))
	}

	// Give up keyboard control to current command
	if !vp.CommandStatus.CaptureInput {
		vp.CommandInput, teaCmd = vp.CommandInput.Update(msg)
		teaCmds = append(teaCmds, teaCmd)
	}

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

	getCmdView := func(withoutTextInput bool) string {
		if withoutTextInput {
			h = 0
		}
		return lipgloss.NewStyle().Foreground(ui.FgColor).Render(
			lipgloss.Place(
				vp.width,
				vp.height-h,
				lipgloss.Left,
				lipgloss.Top,
				cmdView,
			))
	}

	// Do not display text-input when command has exclusive control
	if vp.CommandStatus.CaptureInput {
		return getCmdView(true)
	}

	return lipgloss.JoinVertical(
		lipgloss.Top,
		lipgloss.NewStyle().Foreground(ui.FgColor).Render(
			getCmdView(false),
		),
		cmdrStr,
	)
}

func (vp ViewPort) sendViewportSize() tea.Msg {
	offsetHeight := lipgloss.Height(vp.CommandInput.View())
	if vp.CommandStatus.CaptureInput {
		offsetHeight = 0
	}
	return cmdmodel.ViewportSizeMsg{
		Height: vp.height - offsetHeight,
		Width:  vp.width,
	}
}

func (vp ViewPort) sendStatusMsg(msg string, s cmdmodel.StatusSeverity) func() tea.Msg {
	return func() tea.Msg {
		return cmdmodel.UpdateStatusMsg{
			String:   msg,
			Severity: s,
		}
	}
}
