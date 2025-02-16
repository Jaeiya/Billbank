package cmd

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
you can't add any more commands to that model.
`

type (
	ViewportSizeMsg struct {
		Width  int
		Height int
	}
)

type BaseTree[T any] struct {
	Name     string
	Aliases  []string
	Commands []BaseCommand[T]
}

type BaseCommand[T any] struct {
	Path        string
	NeedArg     bool
	ValidateArg func(arg string) error
	Run         func(T) T
	View        func(T) string
}

type BaseModel[T any] struct {
	cmdMap       map[string]BaseCommand[T]
	cmdTree      BaseTree[T]
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

func NewBaseModel[T any](cmdTree BaseTree[T]) *BaseModel[T] {
	if len(cmdTree.Commands) == 0 {
		logger.LogFatal(
			"Command [%s] has no command paths",
			"Did you forget to add commands to a new command model?",
			cmdTree.Name,
		)
	}

	for _, cmd := range cmdTree.Commands {
		leaves := strings.Split(cmd.Path, " ")
		hasAlias := slices.ContainsFunc(cmdTree.Aliases, func(alias string) bool {
			return leaves[0] == alias
		})

		if hasAlias {
			logger.LogFatal(
				"[%s] command path [%s] does not need to include the command-alias [%s]",
				MsgUsingAliasInCmdPath,
				cmdTree.Name, cmd.Path, leaves[0],
			)
		}
	}

	validateBranches(cmdTree)
	cmdMap := mapCommands(cmdTree)

	logger.Log(
		logger.Info,
		"command [%s] loaded [%d] command paths",
		cmdTree.Name,
		len(cmdTree.Commands),
	)
	logger.Log(
		logger.Debug,
		"[%s] data %+v",
		cmdTree.Name,
		cmdMap,
	)

	return &BaseModel[T]{
		cmdMap:       cmdMap,
		cmdTree:      cmdTree,
		lastCmdError: fmt.Errorf(""),
		isFirstMsg:   true,
	}
}

func (bc *BaseModel[T]) Update(model T, msg tea.Msg) (T, tea.Cmd) {
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

func (bc *BaseModel[T]) View(model T) string {
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

func (m BaseModel[T]) GetViewSize() (int, int) {
	return m.viewWidth, m.viewHeight
}

func (m BaseModel[T]) GetCmdArg() string {
	return m.cmdStatus.Arg
}

func (m BaseModel[T]) GetName() string {
	return m.cmdTree.Name
}

func (m *BaseModel[T]) AddError(err error) {
	m.cmdErrors = append(m.cmdErrors, err)
}

func (m BaseModel[T]) GetErrors() []error {
	return m.cmdErrors
}

func (m *BaseModel[T]) ClearErrors() {
	if len(m.cmdErrors) > 0 {
		m.lastCmdError = fmt.Errorf("")
		m.cmdErrors = nil
	}
}

func (m BaseModel[T]) GetCmdTree() Tree {
	var branches []Branch
	for _, cmd := range m.cmdTree.Commands {
		branches = append(branches, Branch{
			Leaves:      strings.Split(cmd.Path, " "),
			NeedArg:     cmd.NeedArg,
			ValidateArg: cmd.ValidateArg,
		})
	}

	return Tree{
		Aliases:  m.cmdTree.Aliases,
		Branches: branches,
	}
}

func (m BaseModel[T]) IsActivePath(cmdPath string) bool {
	return m.cmdStatus.Path == cmdPath
}

/*
HasView returns true if the current command tree string has
an applicable view associated with it.
*/
func (m BaseModel[T]) HasView() bool {
	cmd := m.cmdMap[m.cmdStatus.Path]
	return cmd.View != nil
}

func (m BaseModel[T]) IsSupported(cmdPath string) bool {
	_, ok := m.cmdMap[cmdPath]
	return ok
}

// IsInitialized checks to make sure that various expected values
// are set.
func (m BaseModel[T]) IsInitialized() bool {
	return len(m.cmdMap) > 0 && len(m.cmdStatus.Path) > 0 && m.viewWidth > 0 &&
		m.viewHeight > 0
}

// Exec executes the current command path in the context of the
// passed model. All detected errors are logged and stored.
func (m *BaseModel[T]) Exec(model T) T {
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

func validateBranches[T any](cmdTree BaseTree[T]) {
	branchMap := map[string]struct{}{}
	for _, cmd := range cmdTree.Commands {
		if cmd.Path == "" && cmd.NeedArg && len(cmdTree.Commands) > 1 {
			logger.LogFatal(
				"[%s] has been initialized as a default command with args, but contains extra commands",
				MsgIsCmdItself,
				cmdTree.Name,
			)
		}

		if cmd.NeedArg && cmd.ValidateArg == nil {
			logger.LogFatal(
				"[%s] command path [%s] is missing an arg validation function.",
				MsgMissingArgFuncErr,
				cmdTree.Name, cmd.Path,
			)
		}
		if _, ok := branchMap[cmd.Path]; ok {
			logger.LogFatal(
				"[%s] contains a duplicate command path [%s]",
				"Is it possible you were testing something and accidentally duplicated a command?",
				cmdTree.Name,
				cmd.Path,
			)
		}

		branchMap[cmd.Path] = struct{}{}
	}
}

func mapCommands[T any](cmdTree BaseTree[T]) map[string]BaseCommand[T] {
	cmdMap := map[string]BaseCommand[T]{}
	for _, cmd := range cmdTree.Commands {
		leaves := strings.Split(cmd.Path, " ")
		path := strings.Join(leaves, " ")

		for _, alias := range cmdTree.Aliases {
			p := strings.TrimSpace(fmt.Sprintf("%s %s", alias, path))
			logger.Log(logger.Hot, "binding command [%s] to [%s] as [%s]", path, alias, p)
			cmdMap[p] = cmd
		}

	}
	return cmdMap
}
