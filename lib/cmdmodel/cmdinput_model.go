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

const MsgDuplicateAliasErr = `
Check to make sure you don't already have a command with
that alias. You may also have accidentally added the
command more than once.`

const MsgInvalidHomeCmdPath = `
Make sure you've entered the entire command path, including the alias.
You also cannot set a home path that requires arguments.

Double check the command paths of the command model you're trying to
access and make sure the path exists.
`

type GoHomeMsg struct{}

type UpdateStatusMsg struct {
	String   string
	Severity StatusSeverity
}

type UpdateCmdMsg struct {
	Model         Interface
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
	homeCmdPath  string
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
func NewInputModel(h *utils.InputHistory, homeCmdPath string, cmdModels ...Model) InputModel {
	inputModel := InputModel{
		aliases:      []string{},
		homeCmdPath:  homeCmdPath,
		CmdHistory:   h,
		CommandInput: commanderInput,
	}

	aliasStore := make(map[string]struct{}, len(cmdModels))

	for _, cmdModel := range cmdModels {
		for _, a := range cmdModel.command.GetCmdData().Aliases {
			if _, ok := aliasStore[a]; ok {
				logger.LogFatal(
					"command alias [%s] already exists",
					MsgDuplicateAliasErr,
					a,
				)
			}
			aliasStore[a] = struct{}{}
			inputModel.aliases = append(inputModel.aliases, a)
		}
		inputModel.commands = append(inputModel.commands, cmdModel)
	}

	if homeCmdPath != "" {
		if _, ok := aliasStore[homeCmdPath]; !ok {
			logger.LogFatal(
				"cannot find home command path [%s] available [%+v]",
				MsgInvalidHomeCmdPath,
				homeCmdPath, aliasStore,
			)
		}
	}

	return inputModel
}

func (m InputModel) Init() tea.Cmd {
	return func() tea.Msg { return GoHomeMsg{} }
}

func (m InputModel) Update(msg tea.Msg) (InputModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		statusStyle = statusStyle.Width(msg.Width)
		m.CommandInput.Width = msg.Width

	case GoHomeMsg:
		if m.homeCmdPath != "" {
			m.CommandInput.SetValue(m.homeCmdPath)
			m, _ = tryParseCmd(m, tea.KeyMsg{})
			m, cmd = tryEnterCmd(m)
			m.CommandInput.Reset()
			return m, cmd
		}

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
				cmdErr := "command parse error:"
				msg := fmt.Sprintf("%s [%s]", cmdErr, m.currCmd.status.Error)
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

	logger.Log(logger.Info, "entering command [%s]", m.CommandInput.Value())
	logger.Log(logger.Debug, "command status [%+v]", cmd.status)

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
			"sending command [%s] status update",
			cmd.status.Path,
		)
		return m, func() tea.Msg { return UpdateCmdMsg{nil, cmd.status} }
	}

	m.lastCmd = cmd
	m.CommandInput.Reset()
	logger.Log(
		logger.Debug,
		"sending [%s] model & status update msg",
		cmd.status.Path,
	)
	teaMsg := UpdateCmdMsg{cmd.command, cmd.status}
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
	for _, cmd := range m.commands {
		logger.Log(
			logger.Insane,
			"test if [%s] is a [%s] command",
			m.CommandInput.Value(),
			cmd.command.GetName(),
		)
		cmdStatus := cmd.ParseCommand(m.CommandInput.Value())
		m.currCmd = cmd
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
