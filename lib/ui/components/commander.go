package components

import (
	"fmt"
	"math/rand"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib/commands"
)

type CommanderModel struct {
	CommandInput textinput.Model
	commands     []commands.Command
	lastCmd      LastCommand
	aliases      []string
	testText     string
	testCount    int
}

type LastCommand struct {
	status commands.CommandResult
	commands.Command
}

type CommanderOption func(*CommanderModel)

func NewCommander(options ...CommanderOption) CommanderModel {
	model := CommanderModel{}
	for _, o := range options {
		o(&model)
	}

	if len(model.commands) == 0 {
		panic("commander requires at least one command")
	}

	model.CommandInput = NewCommanderInput()
	return model
}

func WithCommands(cmds ...commands.Command) CommanderOption {
	return func(m *CommanderModel) {
		aliasStore := map[string]bool{}
		aliases := make([]string, 0, len(aliasStore))
		for _, c := range cmds {
			for _, a := range c.GetStage(0) {
				if _, ok := aliasStore[a]; ok {
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

func (m CommanderModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m CommanderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				m.lastCmd = LastCommand{
					status:  res,
					Command: c,
				}
				if res.IsCommand {
					if !res.IsComplete {
						m.CommandInput.SetSuggestions(res.Suggestions)
					}
					break
				} else {
					m.CommandInput.SetSuggestions(m.aliases)
				}

			}
			return m, cmd
		}
	}
	m.CommandInput, cmd = m.CommandInput.Update(msg)
	return m, cmd
}

func (m CommanderModel) View() string {
	s := fmt.Sprintf("%s\n%s", m.testText, m.CommandInput.View())
	return s
}
