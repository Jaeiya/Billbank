package components

import (
	"fmt"
	"math/rand"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib/commands"
)

type CmdInputModel struct {
	CommandInput textinput.Model
	commands     []commands.Command
	lastCmd      ParsedCmd
	aliases      []string
	testText     string
	testCount    int
}

type ParsedCmd struct {
	status commands.CommandStatus
	commands.Command
}

type CmdInputOption func(*CmdInputModel)

func NewCmdInput(options ...CmdInputOption) CmdInputModel {
	model := CmdInputModel{}
	for _, o := range options {
		o(&model)
	}

	if len(model.commands) == 0 {
		panic("commander requires at least one command")
	}

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
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case " ":
			if len(m.CommandInput.Value()) > 0 {
				lastChar := m.CommandInput.Value()[len(m.CommandInput.Value())-1]
				// Prevent accidental spaces (no valid input will accept consecutive spaces)
				if lastChar == ' ' {
					return m, nil
				}
			}

		case "enter":
			if m.lastCmd.status.IsComplete {
				n := rand.Intn(1000) + 1
				if m.lastCmd.status.Error != nil {
					m.testText = m.lastCmd.status.Error.Error()
				} else {
					m.testText = fmt.Sprintf("Executing Command %d", n)
					m.CommandInput.Reset()
					m.lastCmd = ParsedCmd{}
				}
			} else {
				m.testText = fmt.Sprintf("%v", m.lastCmd)
			}

		default:
			// Use command key validation to restrict user input
			if len(msg.String()) == 1 {
				char := rune(msg.String()[0])
				if m.lastCmd.status.IsComplete {
					if !m.lastCmd.ValidateKey(char) {
						return m, nil
					}
				}
			}

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
	}
	m.CommandInput, cmd = m.CommandInput.Update(msg)
	return m, cmd
}

func (m CmdInputModel) View() string {
	s := fmt.Sprintf("%s\n%s", m.testText, m.CommandInput.View())
	return s
}
