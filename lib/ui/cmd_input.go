package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jaeiya/billbank/lib"
	"github.com/jaeiya/billbank/lib/utils"
	"github.com/jaeiya/billbank/lib/utils/logger"
)

type UpdateStatusMsg struct {
	String   string
	Severity StatusSeverity
}

type StatusSeverity int

const (
	LOW = StatusSeverity(iota)
	MED
	HIGH
)

type CmdInputModel struct {
	CommandInput textinput.Model
	CmdHistory   *utils.InputHistory
	commands     []Command
	currCmd      Command
	lastCmd      Command
	aliases      []string
	statusText   string
}

type CmdInputOption func(*CmdInputModel)

var statusStyle = lipgloss.NewStyle().
	Width(100).
	PaddingLeft(1).
	Background(lib.BgDarkColor).
	Foreground(lib.FgColor)

// TODO - Use an interface to define input history methods
func NewCmdInput(h *utils.InputHistory, options ...CmdInputOption) CmdInputModel {
	model := CmdInputModel{
		aliases: []string{},
	}
	model.CmdHistory = h
	model.CommandInput = NewCommanderInput()

	for _, addCmd := range options {
		addCmd(&model)
	}

	return model
}

func WithCommands(cmds ...Command) CmdInputOption {
	return func(m *CmdInputModel) {
		aliasStore := map[string]bool{}

		for _, cmd := range cmds {
			for _, a := range cmd.tree[0] {
				if aliasStore[a] {
					panic("command alias already exists")
				}
				aliasStore[a] = true
				m.aliases = append(m.aliases, a)
			}
		}

		m.commands = append(m.commands, cmds...)
	}
}

func (m CmdInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m CmdInputModel) Update(msg tea.Msg) (CmdInputModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		statusStyle = statusStyle.Width(msg.Width)
		m.CommandInput.Width = msg.Width

	case UpdateStatusMsg:
		color := lib.FgSuccessColor
		if msg.Severity == MED {
			color = lib.FgWarnColor
		} else if msg.Severity == HIGH {
			color = lib.FgErrColor
		}
		style := statusStyle.Foreground(color)
		m.statusText = style.Render(msg.String)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "up", "down", "alt+j", "alt+k":
			val, reset := m.CmdHistory.Cycle(msg)
			if reset {
				m.CommandInput.Reset()
			} else if len(val) > 0 {
				m.CommandInput.SetValue(val)
				m.CommandInput.CursorEnd()
				return tryParseCmd(m, tea.KeyMsg{})
			}

			return m, nil

		case " ":
			// Prevent accidental spaces
			if isLastCharSpace(m) {
				return m, nil
			}

		case "enter":
			m, cmd = tryEnterCmd(m)
			if m.currCmd.status.Error != nil {
				msg := fmt.Sprintf("%s::[%s]", m.currCmd.status.Error.Error(), m.currCmd.status.BranchStr)
				logger.Log(logger.Attention, fmt.Sprintf("CommandParser: %s", msg))
			}
			cmds = append(cmds, cmd)

		default:
			return onAnyKey(m, msg)
		}
	}
	m.CommandInput, cmd = m.CommandInput.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m CmdInputModel) View() string {
	// Double newline prevents resize artifacts
	s := fmt.Sprintf("%s\n%s", statusStyle.Render(m.statusText), m.CommandInput.View())
	return s
}

func tryEnterCmd(m CmdInputModel) (CmdInputModel, tea.Cmd) {
	m, _ = tryParseCmd(m, tea.KeyMsg{})
	cmd := m.currCmd

	logger.Log(logger.Info, fmt.Sprintf("ExecCommand: [%s]", cmd.status.BranchStr))
	logger.Log(logger.Debug, fmt.Sprintf("CommandStatus: [%+v]", cmd.status))

	if cmd.status.IsCommand {
		if cmd.status.IsComplete && !cmd.status.IsSupported {
			statusStyle = statusStyle.Foreground(lib.FgErrColor)
			m.statusText = "Unsupported Command Chain"
			return m, nil
		} else if !cmd.status.IsComplete {
			statusStyle = statusStyle.Foreground(lib.FgWarnColor)
			m.statusText = "Incomplete Command"
			return m, nil
		}
	}

	if cmd.status.IsComplete && cmd.status.IsCommand {
		if cmd.status.Error != nil {
			m.statusText = cmd.status.Error.Error()
			return m, nil
		}

		statusStyle = statusStyle.Foreground(lib.FgSuccessColor)
		m.statusText = fmt.Sprintf("Executing Command: %s", cmd.status.BranchStr)

		m.CmdHistory.Add(m.CommandInput.Value())

		lastCmd := m.lastCmd
		if lastCmd.GetId() == cmd.GetId() {
			m.CommandInput.Reset()
			return m, func() tea.Msg { return cmd.status }
		}
		m.lastCmd = cmd
		m.CommandInput.Reset()
		model, _ := cmd.model.SetStatus(cmd.status)
		return m, func() tea.Msg { return model }
	}

	statusStyle = statusStyle.Foreground(lib.FgErrColor)
	m.statusText = "Invalid Command"
	return m, nil
}

func onAnyKey(m CmdInputModel, msg tea.KeyMsg) (CmdInputModel, tea.Cmd) {
	m.statusText = ""
	// Restrict user input to "valid" keys
	if len(msg.String()) == 1 {
		char := rune(msg.String()[0])
		if m.currCmd.status.IsComplete {
			if !m.currCmd.ValidateKey(char) {
				return m, nil
			}
		}
	}

	return tryParseCmd(m, msg)
}

func tryParseCmd(m CmdInputModel, msg tea.KeyMsg) (CmdInputModel, tea.Cmd) {
	var cmd tea.Cmd
	m.CommandInput, cmd = m.CommandInput.Update(msg)
	m.currCmd = Command{}
	for _, c := range m.commands {
		cmdStatus := c.ParseCommand(m.CommandInput.Value())
		c.status = cmdStatus
		m.currCmd = c
		if cmdStatus.IsCommand {
			if !cmdStatus.IsComplete {
				m.CommandInput.SetSuggestions(cmdStatus.Branches)
			}
			break
		}
		m.CommandInput.SetSuggestions(m.aliases)
	}
	return m, cmd
}

func isLastCharSpace(m CmdInputModel) bool {
	if len(m.CommandInput.Value()) == 0 {
		return false
	}
	lastChar := m.CommandInput.Value()[len(m.CommandInput.Value())-1]
	return lastChar == ' '
}
