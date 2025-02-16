package cmdmodel

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jaeiya/billbank/lib/logger"
	"github.com/jaeiya/billbank/lib/ui"
	"github.com/jaeiya/billbank/lib/utils"
)

type UpdateStatusMsg struct {
	String   string
	Severity StatusSeverity
}

type UpdateCmdMsg struct {
	Model         ModelCommand
	CommandStatus Status
}

type StatusSeverity int

const (
	LOW = StatusSeverity(iota)
	MED
	HIGH
)

type InputModel struct {
	CommandInput textinput.Model
	CmdHistory   *utils.InputHistory
	commands     []Model
	currCmd      Model
	lastCmd      Model
	aliases      []string
	statusText   string
}

type InputOption func(*InputModel)

var statusStyle = lipgloss.NewStyle().
	Width(100).
	PaddingLeft(1).
	Background(ui.BgDarkColor).
	Foreground(ui.FgColor)

var commanderInput textinput.Model = func() textinput.Model {
	m := textinput.New()
	m.Prompt = ""
	m.PlaceholderStyle = m.PlaceholderStyle.Foreground(lipgloss.Color("#00FFA2"))
	m.CompletionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFA2"))
	m.Focus()
	m.Cursor.Style = m.Cursor.Style.Foreground(lipgloss.Color("#00FFA2"))
	m.PromptStyle = m.Cursor.Style.Foreground(lipgloss.Color("#00FFA2"))
	m.Cursor.BlinkSpeed = time.Millisecond * 500
	m.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFF"))
	m.Prompt = "> "
	m.ShowSuggestions = true
	return m
}()

// TODO - Use an interface to define input history methods
func NewInputModel(h *utils.InputHistory, options ...InputOption) InputModel {
	model := InputModel{
		aliases: []string{},
	}
	model.CmdHistory = h
	model.CommandInput = commanderInput

	for _, addCmd := range options {
		addCmd(&model)
	}

	return model
}

func With(cmds ...Model) InputOption {
	return func(m *InputModel) {
		aliasStore := map[string]bool{}

		for _, cmd := range cmds {
			for _, a := range cmd.model.GetCmdData().Aliases {
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

func (m InputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m InputModel) Update(msg tea.Msg) (InputModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		statusStyle = statusStyle.Width(msg.Width)
		m.CommandInput.Width = msg.Width

	case UpdateStatusMsg:
		color := ui.FgSuccessColor
		if msg.Severity == MED {
			color = ui.FgWarnColor
		} else if msg.Severity == HIGH {
			color = ui.FgErrColor
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
				msg := fmt.Sprintf("%s::[%s]", m.currCmd.status.Error.Error(), m.currCmd.status.Path)
				logger.Log(logger.Attention, "%s", msg)
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

func (m InputModel) View() string {
	s := fmt.Sprintf("%s\n%s", statusStyle.Render(m.statusText), m.CommandInput.View())
	return s
}

func tryEnterCmd(m InputModel) (InputModel, tea.Cmd) {
	// Empty commands will not yet have been parsed.
	if m.CommandInput.Value() == "" {
		m, _ = tryParseCmd(m, tea.KeyMsg{})
	}
	cmd := m.currCmd

	logger.Log(logger.Info, "entering command [%+v]", cmd.status)

	cmdErr := cmd.status.Error
	if cmdErr != nil {
		m.statusText = cmdErr.Error()
		statusStyle = statusStyle.Foreground(ui.FgErrColor)
		if errors.Is(cmdErr, ErrIncompleteCmd) || strings.Contains(cmdErr.Error(), "expected") {
			statusStyle = statusStyle.Foreground(ui.FgWarnColor)
		}
		return m, nil
	}

	statusStyle = statusStyle.Foreground(ui.FgSuccessColor)
	m.statusText = fmt.Sprintf("Executing Command: %s", cmd.status.Path)

	m.CmdHistory.Add(m.CommandInput.Value())

	lastCmd := m.lastCmd
	if lastCmd.GetId() == cmd.GetId() {
		m.CommandInput.Reset()
		logger.Log(
			logger.Debug,
			"sending command [status] update [%s]",
			cmd.status.Path,
		)
		return m, func() tea.Msg { return UpdateCmdMsg{nil, cmd.status} }
	}

	m.lastCmd = cmd
	m.CommandInput.Reset()
	logger.Log(
		logger.Debug,
		"sending command [model & status] update [%s]",
		cmd.status.Path,
	)

	teaMsg := UpdateCmdMsg{cmd.model, cmd.status}
	return m, func() tea.Msg { return teaMsg }
}

func onAnyKey(m InputModel, msg tea.KeyMsg) (InputModel, tea.Cmd) {
	var key string = msg.String()
	if key[0] == 0 {
		key = "ctrl"
	}
	m.statusText = ""
	cmdStatus := m.currCmd.status
	logger.Log(logger.Insane, "[onAnyKey] command status [%+v]", cmdStatus)
	logger.Log(logger.Hot, "[onAnyKey] try parse command on [%s]", key)
	return tryParseCmd(m, msg)
}

func tryParseCmd(m InputModel, msg tea.KeyMsg) (InputModel, tea.Cmd) {
	var cmd tea.Cmd
	m.CommandInput, cmd = m.CommandInput.Update(msg)
	m.currCmd = Model{}
	for _, c := range m.commands {
		logger.Log(
			logger.Insane,
			"test if [%s] is a [%s] command",
			m.CommandInput.Value(),
			c.model.GetName(),
		)
		cmdStatus := c.ParseCommand(m.CommandInput.Value())
		m.currCmd = c
		m.currCmd.status = cmdStatus
		if cmdStatus.IsCommand {
			if errors.Is(cmdStatus.Error, ErrIncompleteCmd) {
				m.CommandInput.SetSuggestions(cmdStatus.PathSuggestions)
			}
			break
		}
		m.CommandInput.SetSuggestions(m.aliases)
	}
	return m, cmd
}

func isLastCharSpace(m InputModel) bool {
	if len(m.CommandInput.Value()) == 0 {
		return false
	}
	lastChar := m.CommandInput.Value()[len(m.CommandInput.Value())-1]
	return lastChar == ' '
}
