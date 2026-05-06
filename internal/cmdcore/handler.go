package cmdcore

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/jaeiya/billbank/internal/logger"
	"github.com/jaeiya/billbank/internal/ui"
	"github.com/jaeiya/billbank/internal/utils"
)

var (
	ErrNotCommand    = fmt.Errorf("unrecognized command")
	ErrIncompleteCmd = fmt.Errorf("incomplete command entry")
	ErrEmptyCommand  = fmt.Errorf("empty command")
)

type CommandHandler interface {
	Update(tea.Msg) (CommandHandler, tea.Cmd)
	View() tea.View
	GetName() string
	GetAliases() []string
	GetId() int
	GetCmdPaths() []string
	ParseCommand(string) CommandState
	GetStatus() CommandState
	SetCmdState(CommandState)
	IsInitialized() bool
}

type (
	HomeMsg struct{}

	StatusBarMsg struct {
		String   string
		Severity StatusSeverity
	}

	ViewportSizeMsg struct {
		Width  int
		Height int
	}

	WindowSizeMsg struct {
		Width  int
		Height int
	}

	UpdateHandlerMsg struct {
		Handler CommandHandler
	}

	ExecCmdMsg struct {
		Width  int
		Height int
	}
)

type CommandState struct {
	Suggestions  []string
	Path         string
	CaptureInput bool
	Error        error
}

type cmdHandler[M any] struct {
	cmdMap   map[string]Command[M]
	name     string
	aliases  []string
	commands []Command[M]
	cmdState CommandState
	cmdModel CommandModel
	errors   struct {
		fatal   error
		command error
	}
	activePath string
	id         int
	viewWidth  int
	viewHeight int
}

func NewCmdHandler[M any](
	name string,
	aliases []string,
	cmds []Command[M],
	cmdModel CommandModel,
) *cmdHandler[M] {
	validateCommands(name, aliases, cmds)

	b := &cmdHandler[M]{
		id:       utils.NewID(),
		name:     name,
		aliases:  aliases,
		commands: cmds,
		cmdModel: cmdModel,
		cmdMap:   make(map[string]Command[M], len(cmds)),
	}

	for _, cmd := range cmds {
		cmdPath := cmd.GetPath()
		leaves := strings.Split(cmdPath, " ")
		path := strings.Join(leaves, " ")
		for _, alias := range aliases {
			fullCmdPath := strings.TrimSpace(fmt.Sprintf("%s %s", alias, path))
			logger.Log(
				logger.Hot,
				"handler [%s] binding command [%s] to [%s] as [%s]",
				name, path, alias, fullCmdPath,
			)
			b.cmdMap[fullCmdPath] = cmd
		}
	}

	logger.Log(logger.Info, "handler [%s] loaded [%d] command paths", name, len(cmds))

	logger.LogFunc(logger.Debug, func() string {
		cmdPaths := make([]string, 0, len(b.cmdMap))
		for key := range b.cmdMap {
			cmdPaths = append(cmdPaths, "["+key+"]")
		}
		return "handler [" + name + "] loaded " + strings.Join(cmdPaths, ", ")
	})

	return b
}

func (ch *cmdHandler[M]) Update(msg tea.Msg) (CommandHandler, tea.Cmd) {
	var teaCmds []tea.Cmd
	var cmd tea.Cmd
	var err error

	switch msg := msg.(type) {
	case ViewportSizeMsg:
		logger.Log(logger.Hot,
			"[%s] handler setting viewport size [%d:%d]",
			ch.name, msg.Width, msg.Height,
		)
		ch.viewHeight = msg.Height
		ch.viewWidth = msg.Width
		if ch.cmdState.Path != "" {
			logger.Log(
				logger.Hot,
				"[%s] handler setting viewport size for command path [%s]",
				ch.name, ch.cmdState.Path,
			)
			ch.cmdModel.SetViewportSize(msg.Width, msg.Height)
		}

	case ExecCmdMsg:
		logger.Log(
			logger.Debug,
			"[%s] handler triggered command execution [%s]",
			ch.name, ch.cmdState.Path,
		)

		if ch.cmdState.Path == "" {
			ch.errors.fatal = fmt.Errorf(
				"[%s] handler tried to execute a command without an active path",
				ch.name,
			)
			logger.Log(logger.Error, ch.errors.fatal.Error())
			return ch, tea.Batch(teaCmds...)
		}

		logger.Log(
			logger.Hot,
			"[%s] handler is updating viewport size [%d:%d]",
			ch.name, msg.Width, msg.Height,
		)
		ch.viewHeight = msg.Height
		ch.viewWidth = msg.Width

		cmdPath := ch.cmdState.Path
		cmd := ch.cmdMap[cmdPath]
		logger.Log(logger.Debug, "[%s] handler is executing command path [%s]", ch.name, cmdPath)
		ch.cmdModel.SetWorkingPath(cmdPath)
		ch.cmdModel.SetViewportSize(msg.Width, msg.Height)
		ch.cmdModel, err = cmd.Run(ch.cmdModel)
		if err != nil {
			if errors.Is(err, ErrModelTypeMismatch) {
				ch.errors.fatal = fmt.Errorf(
					"[%s] handler has passed an invalid model type to command [%s]",
					ch.name, ch.cmdState.Path,
				)
				logger.Log(logger.Error, ch.errors.fatal.Error())
			} else {
				ch.errors.command = fmt.Errorf(
					"[%s] handler got an unexpected error from command [%s]: %w",
					ch.name, ch.cmdState.Path, err,
				)
				logger.Log(logger.Error, ch.errors.command.Error())
			}
		}
		logger.Log(logger.Debug, "[%s] handler has finished executing [%s]", ch.name, ch.cmdState.Path)
	}

	// Do not update an erroring command
	if ch.errors.fatal != nil || ch.errors.command != nil {
		return ch, tea.Batch(teaCmds...)
	}

	ch.cmdModel, cmd = ch.cmdModel.Update(msg)
	teaCmds = append(teaCmds, cmd)
	return ch, tea.Batch(teaCmds...)
}

func (ch cmdHandler[M]) View() tea.View {
	v := tea.NewView("")

	if ch.errors.fatal != nil {
		v.SetContent(ui.NewErrorBox(
			"Fatal Error",
			ch.errors.fatal.Error(),
			ch.viewWidth,
			ch.viewHeight,
		))
		return v
	}

	if ch.errors.command != nil {
		v.SetContent(ui.NewErrorBox(
			"Command Error",
			ch.errors.command.Error(),
			ch.viewWidth,
			ch.viewHeight,
		))
		return v
	}

	cmd, hasPath := ch.cmdMap[ch.cmdState.Path]
	if !hasPath {
		logger.LogFatal(
			"[%s] handler is missing command path [%s]",
			MsgFatalMissingCmd,
			ch.name, ch.cmdState.Path,
		)
	}

	v, err := cmd.View(ch.cmdModel)
	if err != nil {
		if errors.Is(err, ErrModelTypeMismatch) {
			logger.LogFatal(
				"[%s] handler has passed an invalid model to command [%s]",
				"The command handler is to blame.",
				ch.name, ch.cmdState.Path,
			)
		}
		logger.LogFatal(
			"[%s] handler got an unexpected error [%w] from viewing [%s]",
			"Only ModelMismatch errors should be handled here.",
			ch.name, err, ch.cmdState.Path,
		)
	}

	return v
}

// IsInitialized checks to make sure that various expected values
// are set.
func (m cmdHandler[M]) IsInitialized() bool {
	b := len(m.cmdMap) > 0 &&
		len(m.cmdState.Path) > 0 &&
		m.viewWidth > 0 &&
		m.viewHeight > 0
	return b
}

func (m *cmdHandler[M]) SetCmdState(s CommandState) {
	m.cmdState = s
}

func (ch cmdHandler[M]) ParseCommand(input string) CommandState {
	var inputParts []string = strings.Fields(input)
	if len(inputParts) == 0 {
		return CommandState{Error: ErrEmptyCommand}
	}

	alias := inputParts[0]
	inputPathParts := inputParts[1:]
	inputPath := strings.Join(inputPathParts, " ")
	var cmdPaths []string

	if !slices.Contains(ch.aliases, alias) {
		return CommandState{Error: ErrNotCommand}
	}

	cmdState := CommandState{
		Path: alias,
	}

	for _, cmd := range ch.commands {
		cmdPaths = append(cmdPaths, fmt.Sprintf("%s %s", alias, cmd.GetPath()))
		cmdState.Path = strings.TrimSpace(alias + " " + cmd.GetPath())
		cmdState.CaptureInput = cmd.CanCaptureInput()
		pathParts := strings.Split(cmd.GetPath(), " ")

		if cmd.GetArgType() < ArgRequired && inputPath == cmd.GetPath() {
			return cmdState
		}

		hasArg := len(inputPathParts) == len(pathParts)+1

		if cmd.GetArgType() > ArgNone && !hasArg && inputPath == cmd.GetPath() {
			cmdState.Error = fmt.Errorf(
				"expected value after '%s'",
				inputParts[len(inputParts)-1],
			)
			return cmdState
		}

		if cmd.GetArgType() > ArgNone && hasArg {
			argInputPath := strings.Join(inputPathParts[:len(inputPathParts)-1], " ")
			if argInputPath == cmd.GetPath() {
				arg := inputParts[len(inputParts)-1]
				if err := cmd.ResolveArg(arg); err != nil {
					cmdState.Error = err
				}
				return cmdState
			}
		}
	}

	cmdState.Error = ErrIncompleteCmd
	cmdState.Suggestions = populateSuggestions(cmdPaths, inputParts)
	return cmdState
}

func (ch *cmdHandler[M]) SetActivePath(path string) {
	ch.activePath = path
}

func (ch *cmdHandler[M]) SetCommandState(state CommandState) {
	ch.cmdState = state
}

func (ch cmdHandler[M]) ViewSize() (w, h int) {
	return ch.viewWidth, ch.viewHeight
}

func (ch cmdHandler[M]) GetId() int {
	return ch.id
}

func (ch cmdHandler[M]) GetStatus() CommandState {
	return ch.cmdState
}

func (ch cmdHandler[M]) GetName() string {
	return ch.name
}

func (ch cmdHandler[M]) GetAliases() []string {
	aliases := make([]string, len(ch.aliases))
	copy(aliases, ch.aliases)
	return aliases
}

func (ch cmdHandler[M]) GetCmdPaths() (paths []string) {
	paths = make([]string, 0, len(ch.cmdMap))
	for k := range ch.cmdMap {
		paths = append(paths, k)
	}
	return paths
}

func validateCommands[M any](modelName string, aliases []string, cmds []Command[M]) {
	if len(cmds) == 0 {
		logger.LogFatal(
			"Command [%s] has no command paths",
			"Did you forget to add commands to a new command model?",
			modelName,
		)
	}

	cmdMap := map[string]struct{}{}

	for _, cmd := range cmds {
		cmdPath := cmd.GetPath()
		leaves := strings.Split(cmdPath, " ")
		hasAlias := slices.ContainsFunc(aliases, func(alias string) bool {
			return leaves[0] == alias
		})

		if hasAlias {
			logger.LogFatal(
				"command model [%s] already uses the alias [%s]",
				MsgUsingAliasInCmdPathErr,
				modelName, cmdPath,
			)
		}

		if cmdPath == "" && cmd.GetArgType() > ArgNone && len(cmds) > 1 {
			logger.LogFatal(
				"[%s] has been initialized as a default command with args, but also has other command paths",
				MsgIsCmdItselfErr,
				modelName,
			)
		}

		if _, ok := cmdMap[cmdPath]; ok {
			logger.LogFatal(
				"command model [%s] contains a duplicate command path [%s]",
				"Is it possible you were testing something and accidentally duplicated a command?",
				modelName,
				cmdPath,
			)
		}
		cmdMap[cmdPath] = struct{}{}
	}
}

func populateSuggestions(cmdPaths, pathParts []string) []string {
	suggestions := make([]string, 0, 10) // set a reasonable minimum length
	for _, path := range cmdPaths {
		cmdPathParts := strings.Fields(path)
		if len(cmdPathParts) >= len(pathParts) {
			suggestions = append(
				suggestions,
				strings.Join(cmdPathParts[:len(pathParts)], " "),
			)
		}
	}
	return suggestions
}
