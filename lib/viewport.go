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
	hasHiddenInput  bool
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
		logger.Log(logger.Hot, "[WindowSizeMsg] sending viewport size [%d:%d]", vp.width, vp.height)
		teaCmds = append(teaCmds, vp.setupViewport(msg))

	case tea.KeyMsg:
		// Emergency exit
		if msg.String() == "alt+`" {
			return vp, tea.Quit
		}

		if msg.String() == "`" {
			logger.Log(logger.Hot, "[OnGrave] toggling viewport input")
			vp, teaCmds = vp.toggleInput()
			return vp, tea.Batch(teaCmds...)
		}

	case cmdmodel.ReleaseInputMsg:
		teaCmds = append(teaCmds, vp.releaseInput())

	case cmdmodel.UpdateCmdMsg:
		teaCmds = append(teaCmds, vp.updateCommand(msg))

	case CommanderStatusMsg:
		teaCmds = append(teaCmds, vp.sendStatusMsg(msg.String, msg.Severity))
	}

	// Give up keyboard control to current command
	if !vp.hasHiddenInput {
		vp.CommandInput, teaCmd = vp.CommandInput.Update(msg)
		teaCmds = append(teaCmds, teaCmd)
	}

	if vp.CurrentCmdModel != nil {
		_, isKey := msg.(tea.KeyMsg)
		if isKey && vp.hasHiddenInput || !isKey {
			vp.CurrentCmdModel, teaCmd = vp.CurrentCmdModel.Update(msg)
			teaCmds = append(teaCmds, teaCmd)
		}
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
	if vp.hasHiddenInput {
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

func (vp *ViewPort) setupViewport(msg tea.WindowSizeMsg) tea.Cmd {
	vp.height = msg.Height
	vp.width = msg.Width
	return vp.sendViewportSize
}

func (vp *ViewPort) updateCommand(msg cmdmodel.UpdateCmdMsg) tea.Cmd {
	if msg.Model != nil {
		vp.CurrentCmdModel = msg.Model
		logger.Log(logger.Debug, "storing new command model [%s]", msg.CommandStatus.Path)
	}
	vp.CommandStatus = msg.CommandStatus
	if vp.CommandStatus.CaptureInput {
		vp.hasHiddenInput = true
	}
	logger.Log(
		logger.Debug,
		"[UpdateCommand] sending viewport size [%d:%d]",
		vp.width,
		vp.height,
	)
	return vp.sendViewportSize
}

func (vp ViewPort) toggleInput() (ViewPort, []tea.Cmd) {
	vp.hasHiddenInput = !vp.hasHiddenInput
	if !vp.hasHiddenInput {
		return vp, []tea.Cmd{
			func() tea.Msg { return cmdmodel.ReleaseInputMsg{} },
			vp.sendViewportSize,
		}
	}
	return vp, []tea.Cmd{vp.sendViewportSize}
}

func (vp *ViewPort) releaseInput() tea.Cmd {
	vp.hasHiddenInput = false
	vp.CommandStatus.CaptureInput = vp.hasHiddenInput
	return func() tea.Msg {
		return textinput.Blink()
	}
}

func (vp ViewPort) sendViewportSize() tea.Msg {
	offsetHeight := lipgloss.Height(vp.CommandInput.View())
	if vp.hasHiddenInput {
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
