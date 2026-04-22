package internal

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/jaeiya/billbank/internal/cmdmodel"
	"github.com/jaeiya/billbank/internal/logger"
	"github.com/jaeiya/billbank/internal/ui"
)

type (
	CommanderStatusMsg struct {
		String   string
		Severity cmdmodel.StatusSeverity
	}
)

type ViewportSize struct {
	Width  int
	Height int
}

type ViewPort struct {
	cmdInput       cmdmodel.InputModel
	cmdModel       cmdmodel.Interface
	hasHiddenInput bool
	height         int
	width          int
}

func NewViewport(input cmdmodel.InputModel) ViewPort {
	vp := ViewPort{}
	vp.cmdInput = input
	return vp
}

func (vp ViewPort) Init() tea.Cmd {
	teaCmds := []tea.Cmd{
		textinput.Blink,
		vp.cmdInput.Init(),
	}
	logger.Log(logger.Info, "loaded viewport")
	return tea.Batch(teaCmds...)
}

func (vp ViewPort) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var teaCmds []tea.Cmd
	var teaCmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Don't update unless we have a new size
		if msg.Height != vp.height || msg.Width != vp.width {
			vp.height = msg.Height
			vp.width = msg.Width
			logger.Log(logger.Hot, "[WindowSizeMsg] sending viewport size [%d:%d]", vp.width, vp.height)
			teaCmds = append(teaCmds, vp.sendViewportSize(cmdmodel.WindowSizeMsg{}))
		}

	case tea.KeyPressMsg:
		// Emergency exit
		if msg.String() == "alt+`" {
			logger.Log(logger.Debug, "used emergency exit")
			return vp, tea.Quit
		}

		if msg.String() == "`" {
			logger.Log(logger.Hot, "[OnGrave] toggling viewport input")
			vp, teaCmd = vp.toggleInput()
			return vp, teaCmd
		}

	case cmdmodel.UpdateCmdMsg:
		teaCmds = append(teaCmds, vp.updateCommand(msg))

	case CommanderStatusMsg:
		teaCmds = append(teaCmds, vp.sendStatusMsg(msg.String, msg.Severity))
	}

	_, isKey := msg.(tea.KeyPressMsg)

	// Give up keyboard control to current command
	if !vp.hasHiddenInput || !isKey {
		vp.cmdInput, teaCmd = vp.cmdInput.Update(msg)
		teaCmds = append(teaCmds, teaCmd)
	}

	if vp.cmdModel != nil {
		// Ignore key input unless command has exclusive control
		if isKey && vp.hasHiddenInput || !isKey {
			vp.cmdModel, teaCmd = vp.cmdModel.Update(msg)
			teaCmds = append(teaCmds, teaCmd)
		}
	}

	return vp, tea.Batch(teaCmds...)
}

func (vp ViewPort) View() tea.View {
	cmdrStr := vp.cmdInput.View()
	h := lipgloss.Height(cmdrStr.Content)
	v := tea.NewView("")
	v.AltScreen = true
	cmdView := ""

	if vp.cmdModel != nil && vp.cmdModel.IsInitialized() {
		cmdView = vp.cmdModel.View().Content
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
		v.SetContent(getCmdView(true))
		return v
	}

	v.SetContent(lipgloss.JoinVertical(
		lipgloss.Top,
		lipgloss.NewStyle().Foreground(ui.FgColor).Render(
			getCmdView(false),
		),
		cmdrStr.Content,
	))
	v.Cursor = cmdrStr.Cursor
	return v
}

func (vp *ViewPort) updateCommand(msg cmdmodel.UpdateCmdMsg) tea.Cmd {
	if msg.Model != nil {
		vp.cmdModel = msg.Model
		logger.Log(logger.Debug, "storing command model [%s]", msg.Model.GetName())
	}
	if vp.cmdModel.GetStatus().CaptureInput {
		vp.hasHiddenInput = true
	}
	return vp.sendViewportSize(cmdmodel.ViewportSizeMsg{})
}

func (vp ViewPort) getSize() ViewportSize {
	offsetHeight := lipgloss.Height(vp.cmdInput.View().Content)
	if vp.hasHiddenInput {
		offsetHeight = 0
	}
	return ViewportSize{
		Height: vp.height - offsetHeight,
		Width:  vp.width,
	}
}

func (vp ViewPort) toggleInput() (ViewPort, tea.Cmd) {
	var teaCmd tea.Cmd
	vp.hasHiddenInput = !vp.hasHiddenInput
	if !vp.hasHiddenInput {
		teaCmd = textinput.Blink
	}
	// Updating directly, Avoids UI jumping around
	vp.cmdModel, _ = vp.cmdModel.Update(cmdmodel.ViewportSizeMsg(vp.getSize()))
	return vp, teaCmd
}

func (vp ViewPort) sendViewportSize(msgType tea.Msg) tea.Cmd {
	var msg tea.Msg
	vpSize := vp.getSize()
	logger.Log(logger.Hot, "sending viewport size [%d:%d]", vpSize.Width, vpSize.Height)

	switch msgType.(type) {
	case cmdmodel.ViewportSizeMsg:
		msg = cmdmodel.ViewportSizeMsg(vpSize)

	case cmdmodel.WindowSizeMsg:
		msg = cmdmodel.WindowSizeMsg(vpSize)

	default:
		// This should never happen
		panic("invalid viewport message type")
	}

	return func() tea.Msg {
		return msg
	}
}

func (vp ViewPort) sendStatusMsg(msg string, s cmdmodel.StatusSeverity) func() tea.Msg {
	return func() tea.Msg {
		return cmdmodel.StatusBarMsg{
			String:   msg,
			Severity: s,
		}
	}
}
