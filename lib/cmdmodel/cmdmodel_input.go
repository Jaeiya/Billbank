package cmdmodel

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jaeiya/billbank/lib/logger"
	"github.com/jaeiya/billbank/lib/ui"
	"github.com/jaeiya/billbank/lib/utils"
)

type StatusSeverity int

const (
	LOW = StatusSeverity(iota)
	MED
	HIGH
)

type InputModel struct {
	input    textinput.Model
	history  *utils.InputHistory
	homePath string
	commands []Interface
	lastCmd  Interface
	state    struct {
		cmdStatus Status
		activeCmd Interface
	}
	aliases    []string
	statusText string
}

type InputOption func(*InputModel)

var statusStyle = lipgloss.NewStyle().
	Width(100).
	PaddingLeft(1).
	Background(ui.BgDarkColor).
	Foreground(ui.FgColor)

var versionStyle = lipgloss.NewStyle().
	PaddingLeft(1).
	PaddingRight(1).
	Background(lipgloss.Color("#330072")).
	Foreground(ui.BrightMagenta)

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
func NewInputModel(h *utils.InputHistory, homePath string, cmdModels ...Interface) InputModel {
	inputModel := InputModel{
		aliases:  []string{},
		homePath: homePath,
		history:  h,
		input:    commanderInput,
	}

	aliasStore := make(map[string]struct{}, len(cmdModels))
	cmdPaths := []string{}

	for _, cmdModel := range cmdModels {
		for _, alias := range cmdModel.GetAliases() {
			if _, ok := aliasStore[alias]; ok {
				logger.LogFatal(
					"command alias [%s] already exists",
					MsgDuplicateAliasErr,
					alias,
				)
			}
			aliasStore[alias] = struct{}{}
			inputModel.aliases = append(inputModel.aliases, alias)
		}
		cmdPaths = append(cmdPaths, cmdModel.GetCmdPaths()...)
		inputModel.commands = append(inputModel.commands, cmdModel)
	}

	if homePath != "" {
		if !slices.Contains(cmdPaths, homePath) {
			logger.LogFatal(
				"cannot find home command path [%s]",
				MsgInvalidHomeCmdPathErr,
				homePath,
			)
		}
	}

	return inputModel
}

func (m InputModel) Init() tea.Cmd {
	return func() tea.Msg { return HomeMsg{} }
}

func (m InputModel) Update(msg tea.Msg) (InputModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		statusStyle = statusStyle.Width(msg.Width)
		m.input.Width = msg.Width

	case HomeMsg:
		if m.homePath != "" {
			m.input.SetValue(m.homePath)
			m, _ = tryParseCmd(m, tea.KeyMsg{})
			m, cmd = m.tryEnterCmd()
			m.input.Reset()
			return m, cmd
		}

	case StatusBarMsg:
		color := ui.FgSuccessColor
		switch msg.Severity {
		case MED:
			color = ui.FgWarnColor
		case HIGH:
			color = ui.FgErrColor
		}
		style := statusStyle.Foreground(color)
		m.statusText = style.Render(msg.String)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "up", "down", "alt+j", "alt+k":
			val, reset := m.history.Cycle(msg)
			if reset {
				m.input.Reset()
			} else if len(val) > 0 {
				m.input.SetValue(val)
				m.input.CursorEnd()
				return tryParseCmd(m, tea.KeyMsg{})
			}

			return m, nil

		case " ":
			// Prevent accidental spaces
			if isLastCharSpace(m) {
				return m, nil
			}

		case "enter":
			m, cmd = m.onEnter()
			cmds = append(cmds, cmd)

		default:
			return onAnyKey(m, msg)
		}
	}
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m InputModel) View() string {
	version := versionStyle.Render(utils.GetVersion())
	vWidth := lipgloss.Width(version)
	statusWidth := statusStyle.GetWidth()
	status := statusStyle.Width(statusWidth - vWidth).Render(m.statusText)

	line := lipgloss.JoinHorizontal(lipgloss.Left, status, version)

	s := fmt.Sprintf("%s\n%s", line, m.input.View())
	return s
}

func (m InputModel) onEnter() (InputModel, tea.Cmd) {
	var cmd tea.Cmd
	m, cmd = m.tryEnterCmd()
	if m.state.cmdStatus.Error != nil {
		if errors.Is(ErrMisconfiguredArgParser, m.state.cmdStatus.Error) {
			logger.Log(logger.Error,
				"command [%s] path [%s] has a misconfigured arg parser func",
				m.state.activeCmd.GetName(),
				m.state.cmdStatus.Path,
			)
		} else {
			cmdErr := "command parse error:"
			msg := fmt.Sprintf("%s [%s]", cmdErr, m.state.cmdStatus.Error)
			logger.Log(logger.Attention, "%s", msg)
		}
	}
	return m, cmd
}

func (m InputModel) tryEnterCmd() (InputModel, tea.Cmd) {
	// Empty commands will not yet have been parsed.
	if m.input.Value() == "" {
		m, _ = tryParseCmd(m, tea.KeyMsg{})
	}
	cmd := m.state.activeCmd
	cmdStatus := m.state.cmdStatus

	logger.Log(logger.Info, "entering command [%s]", m.input.Value())
	logger.Log(logger.Debug, "command status [%+v]", cmdStatus)

	cmdErr := cmdStatus.Error
	if cmdErr != nil {
		m.statusText = cmdErr.Error()
		statusStyle = statusStyle.Foreground(ui.FgErrColor)
		if errors.Is(cmdErr, ErrIncompleteCmd) || strings.Contains(cmdErr.Error(), "expected") {
			statusStyle = statusStyle.Foreground(ui.FgWarnColor)
		}
		return m, nil
	}

	statusStyle = statusStyle.Foreground(ui.FgSuccessColor)
	m.statusText = fmt.Sprintf("Executing Command: %s", cmdStatus.Path)

	m.history.Add(m.input.Value())

	if m.lastCmd != nil && m.lastCmd.GetId() == cmd.GetId() {
		m.input.Reset()
		logger.Log(
			logger.Debug,
			"sending command [%s] status update",
			cmdStatus.Path,
		)
		cmd.SetStatus(cmdStatus)
		return m, func() tea.Msg { return UpdateCmdMsg{nil} }
	}

	m.lastCmd = cmd
	m.input.Reset()
	logger.Log(
		logger.Debug,
		"sending [%s] model update msg",
		cmdStatus.Path,
	)
	cmd.SetStatus(cmdStatus)
	teaMsg := UpdateCmdMsg{cmd}
	return m, func() tea.Msg { return teaMsg }
}

func onAnyKey(m InputModel, msg tea.KeyMsg) (InputModel, tea.Cmd) {
	var key string = msg.String()
	if key[0] == 0 {
		key = "ctrl"
	}
	m.statusText = ""
	logger.Log(logger.Insane, "[onAnyKey] command status [%+v]", m.state.cmdStatus)
	logger.Log(logger.Hot, "[onAnyKey] try parse command on [%s]", key)
	return tryParseCmd(m, msg)
}

func tryParseCmd(m InputModel, msg tea.KeyMsg) (InputModel, tea.Cmd) {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.state = struct {
		cmdStatus Status
		activeCmd Interface
	}{cmdStatus: Status{}, activeCmd: nil}
	for _, cmd := range m.commands {
		logger.Log(
			logger.Insane,
			"test if [%s] is a [%s] command",
			m.input.Value(),
			cmd.GetName(),
		)
		status := cmd.ParseCommand(m.input.Value())
		m.state.cmdStatus = status
		m.state.activeCmd = cmd
		if status.IsCommand {
			if errors.Is(status.Error, ErrIncompleteCmd) {
				m.input.SetSuggestions(status.Suggestions)
			}
			break
		}
		m.input.SetSuggestions(m.aliases)
	}
	return m, cmd
}

func isLastCharSpace(m InputModel) bool {
	if len(m.input.Value()) == 0 {
		return false
	}
	lastChar := m.input.Value()[len(m.input.Value())-1]
	return lastChar == ' '
}
