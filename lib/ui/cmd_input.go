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
	CommandInput  textinput.Model
	CmdHistory    *utils.InputHistory
	commands      []Command
	currentCmdPos int
	lastCmdPos    int
	aliases       []string
	statusText    string
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
		aliases:    []string{},
		lastCmdPos: -1,
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
			currCmd := m.GetCommand(m.currentCmdPos)
			logger.Log(logger.Info, fmt.Sprintf("ExecCommand: [%s]", currCmd.status.TreeStr))
			m, cmd = tryEnterCmd(m)
			if currCmd.status.Error != nil {
				msg := fmt.Sprintf("%s::[%s]", currCmd.status.Error.Error(), currCmd.status.TreeStr)
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

func (m CmdInputModel) GetCommand(pos int) Command {
	if pos >= 0 && pos < len(m.commands) {
		return m.commands[pos]
	}
	panic("tried to access non-existent command")
}

func tryEnterCmd(m CmdInputModel) (CmdInputModel, tea.Cmd) {
	var currCmd Command = m.GetCommand(m.currentCmdPos)

	if currCmd.status.IsCommand {
		if !currCmd.status.IsSupported {
			statusStyle = statusStyle.Foreground(lib.FgErrColor)
			m.statusText = "Unsupported Command Chain"
			return m, nil
		} else if !currCmd.status.IsComplete {
			statusStyle = statusStyle.Foreground(lib.FgWarnColor)
			m.statusText = "Incomplete Command"
			return m, nil
		}
	}

	if currCmd.status.IsComplete {
		if currCmd.status.Error != nil {
			m.statusText = currCmd.status.Error.Error()
			return m, nil
		}

		statusStyle = statusStyle.Foreground(lib.FgSuccessColor)
		m.statusText = fmt.Sprintf("Executing Command: %s", currCmd.status.TreeStr)

		m.CmdHistory.Add(m.CommandInput.Value())

		var lastCmd Command
		if m.lastCmdPos == -1 {
			lastCmd = Command{}
		} else {
			lastCmd = m.GetCommand(m.lastCmdPos)
		}

		if lastCmd.GetId() == currCmd.GetId() {
			m.CommandInput.Reset()
			return m, func() tea.Msg { return currCmd.status }
		}
		m.lastCmdPos = m.currentCmdPos
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
		currCmd := m.GetCommand(m.currentCmdPos)
		if currCmd.status.IsComplete {
			if !currCmd.ValidateKey(char) {
				return m, nil
			}
		}
	}

	return tryParseCmd(m, msg)
}

func tryParseCmd(m CmdInputModel, msg tea.KeyMsg) (CmdInputModel, tea.Cmd) {
	var cmd tea.Cmd
	m.CommandInput, cmd = m.CommandInput.Update(msg)
	for i, c := range m.commands {
		res := c.ParseCommand(m.CommandInput.Value())
		m.currentCmdPos = i
		m.commands[i].status = res
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
