package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jaeiya/billbank/lib/commands"
)

type TestMsg bool

type CmdInputModel struct {
	CommandInput textinput.Model
	CmdHistory   *CmdHistory
	commands     []commands.Command
	lastCmd      ParsedCmd
	aliases      []string
	statusText   string
}

type ParsedCmd struct {
	status commands.CommandStatus
	commands.Command
}

type CmdInputOption func(*CmdInputModel)

var statusStyle = lipgloss.NewStyle().
	Width(100).
	PaddingLeft(1).
	Background(lipgloss.Color("#151718")).
	Foreground(lipgloss.Color("#FFA200"))

var (
	okColor   = lipgloss.Color("#00FFA2")
	warnColor = lipgloss.Color("#FFA200")
	errColor  = lipgloss.Color("#FF5FC5")
)

func NewCmdInput(options ...CmdInputOption) CmdInputModel {
	model := CmdInputModel{}
	for _, o := range options {
		o(&model)
	}

	if len(model.commands) == 0 {
		panic("commander requires at least one command")
	}

	model.CmdHistory = NewCmdHistory()
	model.CommandInput = NewCommanderInput()
	return model
}

func WithCommands(cmds ...commands.Command) CmdInputOption {
	return func(m *CmdInputModel) {
		aliasStore := map[string]bool{}
		var aliases []string

		for _, cmd := range cmds {
			for _, a := range cmd.GetAliases() {
				if aliasStore[a] {
					panic("command alias already exists")
				}
				aliasStore[a] = true
				aliases = append(aliases, a)
			}
		}

		m.aliases = aliases
		m.commands = cmds
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

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "up", "down", "alt+j", "alt+k":
			val, reset := m.CmdHistory.Cycle(msg)
			if reset {
				m.CommandInput.Reset()
				m.lastCmd = ParsedCmd{}
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
			// cmds = append(cmds, cmd)
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
	if m.lastCmd.status.IsComplete {
		if m.lastCmd.status.Error != nil {
			m.statusText = m.lastCmd.status.Error.Error()
			return m, nil
		}
		statusStyle = statusStyle.Foreground(okColor)
		m.statusText = fmt.Sprintf("Executing Command %s", m.lastCmd.status.Arg)
		m.CmdHistory.Add(m.CommandInput.Value())
		m.CommandInput.Reset()
		m.lastCmd = ParsedCmd{}
		return m, func() tea.Msg { return TestMsg(true) }
	}

	if m.lastCmd.status.IsCommand && !m.lastCmd.status.IsComplete {
		statusStyle = statusStyle.Foreground(warnColor)
		m.statusText = "Incomplete Command"
		return m, nil
	}

	statusStyle = statusStyle.Foreground(errColor)
	m.statusText = "Invalid Command"
	return m, nil
}

func onAnyKey(m CmdInputModel, msg tea.KeyMsg) (CmdInputModel, tea.Cmd) {
	m.statusText = ""
	// Restrict user input to "valid" keys
	if len(msg.String()) == 1 {
		char := rune(msg.String()[0])
		if m.lastCmd.status.IsComplete {
			if !m.lastCmd.ValidateKey(char) {
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
		m.lastCmd = ParsedCmd{
			status:  res,
			Command: c,
		}
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
