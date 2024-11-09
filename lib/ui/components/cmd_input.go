package components

import (
	"fmt"
	"math/rand"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jaeiya/billbank/lib/commands"
)

type CmdInputModel struct {
	CommandInput  textinput.Model
	commands      []commands.Command
	lastCmd       ParsedCmd
	aliases       []string
	statusText    string
	cmdHistory    []string
	cmdHistoryPos int
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

	model.cmdHistoryPos = -1
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

func (m CmdInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		statusStyle = statusStyle.Width(msg.Width)
		m.CommandInput.Width = msg.Width

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "up", "down", "alt+j", "alt+k":
			return handleCmdHistory(m, msg)

		case " ":
			// Prevent accidental spaces
			if isDoubleSpace(m) {
				return m, nil
			}

		case "enter":
			m = tryEnterCmd(m)

		default:
			return onAnyKey(m, msg)
		}
	}
	m.CommandInput, cmd = m.CommandInput.Update(msg)
	return m, cmd
}

func (m CmdInputModel) View() string {
	// Double newline prevents resize artifacts
	s := fmt.Sprintf("%s\n%s\n\n", statusStyle.Render(m.statusText), m.CommandInput.View())
	return s
}

func tryEnterCmd(m CmdInputModel) CmdInputModel {
	if m.lastCmd.status.IsComplete {
		n := rand.Intn(1000) + 1
		if m.lastCmd.status.Error != nil {
			m.statusText = m.lastCmd.status.Error.Error()
			return m
		}
		statusStyle = statusStyle.Foreground(okColor)
		m.statusText = fmt.Sprintf("Executing Command %d", n)
		m.cmdHistory = append([]string{m.CommandInput.Value()}, m.cmdHistory...)
		m.CommandInput.Reset()
		m.lastCmd = ParsedCmd{}
		m.cmdHistoryPos = -1
		return m
	}

	if m.lastCmd.status.IsCommand && !m.lastCmd.status.IsComplete {
		statusStyle = statusStyle.Foreground(warnColor)
		m.statusText = "Incomplete Command"
		return m
	}

	statusStyle = statusStyle.Foreground(errColor)
	m.statusText = "Invalid Command"
	return m
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

func isDoubleSpace(m CmdInputModel) bool {
	if len(m.CommandInput.Value()) > 0 {
		lastChar := m.CommandInput.Value()[len(m.CommandInput.Value())-1]
		if lastChar == ' ' {
			return true
		}
	}
	return false
}

func handleCmdHistory(m CmdInputModel, msg tea.KeyMsg) (CmdInputModel, tea.Cmd) {
	if len(m.cmdHistory) == 0 {
		return m, nil
	}

	switch msg.String() {
	case "up", "alt+k":
		if m.cmdHistoryPos+1 < len(m.cmdHistory) {
			m.cmdHistoryPos++
			m.CommandInput.SetValue(m.cmdHistory[m.cmdHistoryPos])
			m.CommandInput.CursorEnd()
		}

	case "down", "alt+j":
		if m.cmdHistoryPos > 0 {
			m.cmdHistoryPos--
			m.CommandInput.SetValue(m.cmdHistory[m.cmdHistoryPos])
			m.CommandInput.CursorEnd()
		} else {
			m.cmdHistoryPos = -1
			m.CommandInput.Reset()
		}
	}

	return tryParseCmd(m, msg)
}
