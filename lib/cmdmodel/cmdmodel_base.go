package cmdmodel

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib/logger"
	"github.com/jaeiya/billbank/lib/ui"
)

const MsgMissingArgFuncErr = `
If a command path requires an argument, then it should also validate
that argument. It's recommended to validate the arg inside this
function because it can notify the user on error.`

const MsgIsCmdItself = `
Models with default commands (commands executed just by their alias),
that take arguments, do not support extra commands. For instance, if
your model has a "view" alias that takes an argument for the kind of
view to display:

[view 1] or [view 2]

You're limited to just a single command path, the default path. You
cannot then create more commands that take a more specific view type
like so:

[view details] or [view list]

The above second-order commands will be ignored as if they don't exist.
Either setup your model to accept args directly or commands, but not
both. You can also setup your model to execute a default command without
args, allowing you to add extra commands.
`

const MsgUsingAliasInCmdPath = `
Command-paths do not need to include the alias of their parent model.

If you're trying to setup a default command (a command executed by its
aliases alone) then just add a command with an empty string for its path
like so:

{ Path: "", Run: loadCmd, View: loadView }

Where loadCmd and loadView are functions that take your command model
as an argument. This will allow the execution of the model aliases,
as if they were commands themselves.

Be aware though, that if you set an empty path's "NeedArg" to true,
you can't add any more commands to that model.`

type ArgType int

const (
	ArgNone = ArgType(iota)
	ArgOptional
	ArgRequired
)

type (
	ViewportSizeMsg struct {
		Width  int
		Height int
	}
)

type BaseCmdData[T any] struct {
	Name     string
	Aliases  []string
	Commands []BaseCommand[T]
}

type BaseCommand[T any] struct {
	Path        string
	ArgType     ArgType
	ValidateArg func(arg string) error
	Run         func(T) T
	View        func(T) string
}

type ModelBase[T any] struct {
	cmdMap       map[string]BaseCommand[T]
	cmdData      BaseCmdData[T]
	cmdStatus    Status
	cmdErrors    []error
	lastCmdError error
	isFirstMsg   bool
	viewWidth    int
	viewHeight   int
	hasStaleView bool
	staleView    string
	stalePath    string
	// Whether or not a tea.Msg is an interrupt which
	// we'll use to prevent things like log spamming.
	isInterrupt bool
}

func NewModelBase[T any](cmdData BaseCmdData[T]) *ModelBase[T] {
	validateCmdData(cmdData)

	cmdMap := map[string]BaseCommand[T]{}
	for _, cmd := range cmdData.Commands {
		leaves := strings.Split(cmd.Path, " ")
		path := strings.Join(leaves, " ")
		for _, alias := range cmdData.Aliases {
			fullCmdPath := strings.TrimSpace(fmt.Sprintf("%s %s", alias, path))
			logger.Log(
				logger.Hot,
				"binding command [%s] to [%s] as [%s]",
				path, alias, fullCmdPath,
			)
			cmdMap[fullCmdPath] = cmd
		}

	}

	logger.Log(
		logger.Info,
		"command [%s] loaded [%d] command paths",
		cmdData.Name,
		len(cmdData.Commands),
	)

	logger.LogFunc(logger.Debug, func() string {
		var cmdPaths []string = make([]string, 0, len(cmdMap))
		for key := range cmdMap {
			cmdPaths = append(cmdPaths, fmt.Sprintf("[%s]", key))
		}
		pathStrings := fmt.Sprintf("%+v", strings.Join(cmdPaths, ", "))
		return fmt.Sprintf("command [%s] loaded %s", cmdData.Name, pathStrings)
	})

	return &ModelBase[T]{
		cmdMap:       cmdMap,
		cmdData:      cmdData,
		lastCmdError: fmt.Errorf(""),
		isFirstMsg:   true,
	}
}

func (bc *ModelBase[T]) Update(model T, msg tea.Msg) (T, tea.Cmd) {
	var teaCmds []tea.Cmd
	defer func() { bc.isFirstMsg = false }()

	switch msg := msg.(type) {
	case ViewportSizeMsg:
		logger.Log(logger.Hot, "setting viewport size [%dx%d]", msg.Width, msg.Height)
		bc.viewHeight = msg.Height
		bc.viewWidth = msg.Width
		bc.hasStaleView = false
		model = bc.Exec(model)

	case UpdateCmdMsg:
		// The command path won't be executed until the next model update
		// therefore we need to mark the view as stale, so it won't
		// try to view uninitialized model data.
		oldPath := bc.cmdStatus.Path
		bc.hasStaleView = true
		bc.cmdStatus = msg.CommandStatus
		logger.Log(logger.Debug,
			"updated command path [%s] to [%s]",
			oldPath,
			bc.cmdStatus.Path,
		)

	case cursor.BlinkMsg, tea.MouseMsg:
		bc.isInterrupt = true
	}

	return model, tea.Batch(teaCmds...)
}

func (bc *ModelBase[T]) View(model T) string {
	cmdPath := bc.cmdStatus.Path
	cmd := bc.cmdMap[cmdPath]

	if bc.hasStaleView {
		logger.Log(logger.Hot, "loading [stale] view [%s]", bc.stalePath)
		return bc.staleView
	}

	if !bc.isInterrupt {
		logger.Log(
			logger.Hot,
			"loading [current] view [%s]",
			bc.cmdStatus.Path,
		)
	}

	errs := bc.GetErrors()
	if len(errs) > 0 {
		return ui.NewErrorBox(
			"Command Error",
			errs[0].Error(),
			bc.viewWidth,
			bc.viewHeight,
		)
	}

	bc.staleView = cmd.View(model)
	bc.stalePath = bc.cmdStatus.Path
	return cmd.View(model)
}

func (m ModelBase[T]) GetViewSize() (int, int) {
	return m.viewWidth, m.viewHeight
}

func (m ModelBase[T]) GetCmdArg() string {
	return m.cmdStatus.Arg
}

func (m ModelBase[T]) GetName() string {
	return m.cmdData.Name
}

func (m *ModelBase[T]) AddError(err error) {
	m.cmdErrors = append(m.cmdErrors, err)
}

func (m ModelBase[T]) GetErrors() []error {
	return m.cmdErrors
}

func (m *ModelBase[T]) ClearErrors() {
	if len(m.cmdErrors) > 0 {
		m.lastCmdError = fmt.Errorf("")
		m.cmdErrors = nil
	}
}

func (m ModelBase[T]) GetCmdData() CommandData {
	var commands []ModelCommand
	for _, cmd := range m.cmdData.Commands {
		commands = append(commands, newModelCommand(cmd))
	}

	return CommandData{
		Aliases:  m.cmdData.Aliases,
		Commands: commands,
	}
}

func (m ModelBase[T]) IsActivePath(cmdPath string) bool {
	return m.cmdStatus.Path == cmdPath
}

func (m ModelBase[T]) IsSupported(cmdPath string) bool {
	_, ok := m.cmdMap[cmdPath]
	return ok
}

// IsInitialized checks to make sure that various expected values
// are set.
func (m ModelBase[T]) IsInitialized() bool {
	return len(m.cmdMap) > 0 && len(m.cmdStatus.Path) > 0 && m.viewWidth > 0 &&
		m.viewHeight > 0
}

// Exec executes the current command path in the context of the
// passed model. All detected errors are logged and stored.
func (m *ModelBase[T]) Exec(model T) T {
	m.isInterrupt = false
	cmdPath := m.cmdStatus.Path
	cmd := m.cmdMap[cmdPath]

	m.ClearErrors()

	logger.Log(logger.Hot, "executing command path [%s]", cmdPath)

	if cmd.Run == nil {
		err := fmt.Errorf("[%s] tried to execute missing implementation func()", cmdPath)
		if m.lastCmdError.Error() != err.Error() {
			logger.Log(logger.Error, "%s", err)
			m.lastCmdError = err
			m.AddError(err)
		}
		return model
	}

	if cmd.View == nil {
		err := fmt.Errorf("[%s] tried to execute missing view func()", cmdPath)
		if m.lastCmdError.Error() != err.Error() {
			logger.Log(logger.Error, "%s", err)
			m.lastCmdError = err
			m.AddError(err)
		}
		return model
	}

	model = cmd.Run(model)
	errs := m.GetErrors()
	if len(errs) > 0 {
		if m.lastCmdError.Error() != errs[0].Error() {
			logger.Log(logger.Error, "%s", errs[0].Error())
		}
		m.lastCmdError = errs[0]
	}

	return model
}

func validateCmdData[T any](cmdData BaseCmdData[T]) {
	if len(cmdData.Commands) == 0 {
		logger.LogFatal(
			"Command [%s] has no command paths",
			"Did you forget to add commands to a new command model?",
			cmdData.Name,
		)
	}

	for _, cmd := range cmdData.Commands {
		leaves := strings.Split(cmd.Path, " ")
		hasAlias := slices.ContainsFunc(cmdData.Aliases, func(alias string) bool {
			return leaves[0] == alias
		})

		if hasAlias {
			logger.LogFatal(
				"[%s] command path [%s] does not need to include the command-alias [%s]",
				MsgUsingAliasInCmdPath,
				cmdData.Name, cmd.Path, leaves[0],
			)
		}
	}

	cmdMap := map[string]struct{}{}
	for _, cmd := range cmdData.Commands {
		if cmd.Path == "" && cmd.ArgType == ArgRequired && len(cmdData.Commands) > 1 {
			logger.LogFatal(
				"[%s] has been initialized as a default command with args, but contains extra commands",
				MsgIsCmdItself,
				cmdData.Name,
			)
		}

		if cmd.ArgType > ArgNone && cmd.ValidateArg == nil {
			logger.LogFatal(
				"[%s] command path [%s] is missing an arg validation function.",
				MsgMissingArgFuncErr,
				cmdData.Name, cmd.Path,
			)
		}
		if _, ok := cmdMap[cmd.Path]; ok {
			logger.LogFatal(
				"[%s] contains a duplicate command path [%s]",
				"Is it possible you were testing something and accidentally duplicated a command?",
				cmdData.Name,
				cmd.Path,
			)
		}
		cmdMap[cmd.Path] = struct{}{}
	}
}
