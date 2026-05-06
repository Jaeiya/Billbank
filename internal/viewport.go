package internal

import (
	"fmt"
	"reflect"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/jaeiya/billbank/internal/cmdcore"
	"github.com/jaeiya/billbank/internal/logger"
	"github.com/jaeiya/billbank/internal/ui"
)

type (
	CommanderStatusMsg struct {
		String   string
		Severity cmdcore.StatusSeverity
	}
)

type ViewportSize struct {
	Width  int
	Height int
}

type ViewPort struct {
	cmdInput       cmdcore.InputModel
	cmdHandler     cmdcore.CommandHandler
	hasHiddenInput bool
	winHeight      int
	winWidth       int
}

func NewViewport(input cmdcore.InputModel) ViewPort {
	vp := ViewPort{}
	vp.cmdInput = input
	return vp
}

func (vp ViewPort) Init() tea.Cmd {
	teaCmds := []tea.Cmd{
		textinput.Blink,
		vp.cmdInput.Init(),
	}
	logger.Log(logger.Info, "viewport loaded")
	return tea.Batch(teaCmds...)
}

func (vp ViewPort) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var teaCmds []tea.Cmd
	var teaCmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Don't update unless we have a new size
		if msg.Height != vp.winHeight || msg.Width != vp.winWidth {
			logger.Log(logger.Hot, "setting window size [%d:%d]", msg.Width, msg.Height)
			vp.winHeight = msg.Height
			vp.winWidth = msg.Width
			teaCmds = append(teaCmds, vp.sendViewportSize())
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

	case cmdcore.UpdateHandlerMsg:
		teaCmds = append(teaCmds, vp.updateHandler(msg))
		// We don't need to propagate this message
		return vp, tea.Batch(teaCmds...)

	case CommanderStatusMsg:
		teaCmds = append(teaCmds, vp.sendStatusMsg(msg.String, msg.Severity))
	}

	_, isKey := msg.(tea.KeyPressMsg)

	// Give up keyboard control to current command
	if !vp.hasHiddenInput || !isKey {
		logger.LogFunc(logger.Insane, func() string {
			return fmt.Sprintf("sending [%+v] to command input", reflect.TypeOf(msg))
		})
		vp.cmdInput, teaCmd = vp.cmdInput.Update(msg)
		teaCmds = append(teaCmds, teaCmd)
	}

	if vp.cmdHandler != nil {
		// Ignore key input unless command has exclusive control
		if isKey && vp.hasHiddenInput || !isKey {
			logger.LogFunc(logger.Hot, func() string {
				return fmt.Sprintf("sending [%+v] to [%s] handler",
					reflect.TypeOf(msg), vp.cmdHandler.Name(),
				)
			})
			vp.cmdHandler, teaCmd = vp.cmdHandler.Update(msg)
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

	if vp.cmdHandler != nil && vp.cmdHandler.IsInitialized() {
		cmdView = vp.cmdHandler.View().Content
	}

	getCmdView := func(withoutTextInput bool) string {
		if withoutTextInput {
			h = 0
		}
		return lipgloss.NewStyle().Foreground(ui.FgColor).Render(
			lipgloss.Place(
				vp.winWidth,
				vp.winHeight-h,
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

func (vp *ViewPort) updateHandler(msg cmdcore.UpdateHandlerMsg) tea.Cmd {
	if msg.Handler != nil {
		vp.cmdHandler = msg.Handler
		logger.Log(
			logger.Debug,
			"storing [%s] handler",
			msg.Handler.Name(),
		)
	}

	if vp.cmdHandler.CommandState().CaptureInput {
		vp.hasHiddenInput = true
	}

	return func() tea.Msg {
		return cmdcore.ExecCmdMsg(vp.getSize())
	}
}

func (vp ViewPort) getSize() ViewportSize {
	offsetHeight := lipgloss.Height(vp.cmdInput.View().Content)
	if vp.hasHiddenInput {
		offsetHeight = 0
	}
	return ViewportSize{
		Height: vp.winHeight - offsetHeight,
		Width:  vp.winWidth,
	}
}

func (vp ViewPort) toggleInput() (ViewPort, tea.Cmd) {
	var teaCmd tea.Cmd
	vp.hasHiddenInput = !vp.hasHiddenInput
	if !vp.hasHiddenInput {
		teaCmd = textinput.Blink
	}
	// Updating directly, Avoids UI jumping around
	vp.cmdHandler, _ = vp.cmdHandler.Update(cmdcore.ViewportSizeMsg(vp.getSize()))
	return vp, teaCmd
}

func (vp ViewPort) sendViewportSize() tea.Cmd {
	vpSize := vp.getSize()
	logger.Log(
		logger.Hot,
		"sending viewport size msg [%d:%d]",
		vpSize.Width,
		vpSize.Height,
	)

	return func() tea.Msg {
		return cmdcore.ViewportSizeMsg(vpSize)
	}
}

func (vp ViewPort) sendStatusMsg(msg string, s cmdcore.StatusSeverity) func() tea.Msg {
	return func() tea.Msg {
		return cmdcore.StatusBarMsg{
			String:   msg,
			Severity: s,
		}
	}
}
