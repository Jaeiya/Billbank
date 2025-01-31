package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jaeiya/billbank/lib"
	"github.com/jaeiya/billbank/lib/utils"
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
	currentCmd   Command
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

	// options = append(options, WithCommands(NewDebugCmd(model.CmdHistory)))
	for _, addCmd := range options {
		addCmd(&model)
	}

	return model
}

func WithCommands(cmds ...Command) CmdInputOption {
	return func(m *CmdInputModel) {
		aliasStore := map[string]bool{}

		for _, cmd := range cmds {
			for _, a := range cmd.GetAliases() {
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
				m.currentCmd = Command{}
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
			utils.Log(utils.Info, fmt.Sprintf("ExecCommand: [%s]", m.currentCmd.status.TreeStr))
			m, cmd = tryEnterCmd(m)
			if m.currentCmd.status.Error != nil {
				utils.Log(utils.Attention, fmt.Sprintf("CommandError: %s", m.currentCmd.status.Error.Error()))
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
	if m.currentCmd.status.IsCommand {
		if !m.currentCmd.status.IsSupported {
			statusStyle = statusStyle.Foreground(lib.FgErrColor)
			m.statusText = "Unsupported Command Chain"
			return m, nil
		} else if !m.currentCmd.status.IsComplete {
			statusStyle = statusStyle.Foreground(lib.FgWarnColor)
			m.statusText = "Incomplete Command"
			return m, nil
		}
	}

	if m.currentCmd.status.IsComplete {
		if m.currentCmd.status.Error != nil {
			m.statusText = m.currentCmd.status.Error.Error()
			return m, nil
		}

		statusStyle = statusStyle.Foreground(lib.FgSuccessColor)
		m.statusText = fmt.Sprintf("Executing Command: %s", m.currentCmd.status.TreeStr)

		m.CmdHistory.Add(m.CommandInput.Value())
		currCmd := m.currentCmd

		if m.lastCmd.GetId() == currCmd.GetId() {
			m.currentCmd = Command{}
			m.CommandInput.Reset()
			return m, func() tea.Msg { return currCmd.status }
		}
		m.lastCmd = m.currentCmd
		m.currentCmd = Command{}
		m.CommandInput.Reset()
		return m, func() tea.Msg { return currCmd.model.SetStatus(currCmd.status) }
	}

	statusStyle = statusStyle.Foreground(lib.FgErrColor)

	// Force cmd to set error state if user enters an empty
	// command. Parser only activates when the user types.
	if m.CommandInput.Value() == "" {
		m, _ = tryParseCmd(m, tea.KeyMsg{})
	}
	m.statusText = "Invalid Command"
	return m, nil
}

func onAnyKey(m CmdInputModel, msg tea.KeyMsg) (CmdInputModel, tea.Cmd) {
	m.statusText = ""
	// Restrict user input to "valid" keys
	if len(msg.String()) == 1 {
		char := rune(msg.String()[0])
		if m.currentCmd.status.IsComplete {
			if !m.currentCmd.ValidateKey(char) {
				return m, nil
			}
		}
	}

	return tryParseCmd(m, msg)
}

func tryParseCmd(m CmdInputModel, msg tea.KeyMsg) (CmdInputModel, tea.Cmd) {
	var cmd tea.Cmd
	m.CommandInput, cmd = m.CommandInput.Update(msg)
	for _, c := range m.commands {
		res := c.ParseCommand(m.CommandInput.Value())
		m.currentCmd = c
		m.currentCmd.status = res
		if res.IsCommand {
			if !res.IsComplete {
				m.CommandInput.SetSuggestions(res.Suggestions)
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
