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
If a branch requires an argument, then it also requires validation. If you
have not included a validation function, then you're not validating the
users input, which is an anti-pattern.
`

const MsgIsCmdItself = `
Commands that are intended to be used by themselves (called by just
their alias), with arguments, cannot also contain other branch commands.
For instance, if you have a view command that takes an argument for the
kind of view to display:

[view 1] or [view 2]

You cannot also support branch commands, which would also qualify
as arguments passed to the view command:

[view details] or [view list]

Either set up your command to accept arguments or other command branches,
but not both.
`

type (
	ViewportSizeMsg struct {
		Width  int
		Height int
	}
)

type BranchCommand[T any] struct {
	Path string
	Run  func(T) T
	View func(T) string
}

type BaseModel[T any] struct {
	cmdMap       map[string]BranchCommand[T]
	cmdTree      Tree
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

func NewBaseModel[T any](cmdTree Tree, cmds []BranchCommand[T]) *BaseModel[T] {
	if len(cmdTree.Branches) == 0 {
		logger.LogFatal(
			"[%s] has an empty command-branch list",
			"All commands require at least one command branch.",
			cmdTree.Name,
		)
	}

	if len(cmds) == 0 {
		logger.LogFatal(
			"[%s] has an empty branch-command list",
			"All commands require at least one branch command.",
			cmdTree.Name,
		)
	}

	if len(cmds) != len(cmdTree.Branches) {
		logger.LogFatal(
			"[%s] has an inconsistent number of command-branches and branch-commands",
			"Commands should always have a matching branch-command with a command-branch",
			cmdTree.Name,
		)
	}

	for _, cmd := range cmds {
		leaves := strings.Split(cmd.Path, " ")
		hasAlias := slices.ContainsFunc(cmdTree.Aliases, func(alias string) bool {
			return leaves[0] == alias
		})

		if hasAlias {
			logger.LogFatal(
				"Branch [%s] does not need to include [%s] as part of the command path",
				"Command paths do not need to include the alias of the command, as part of the path.",
				cmd.Path,
				leaves[0],
			)
		}
	}

	validateBranches(cmds, cmdTree)
	cmdMap := mapCommands(cmds, cmdTree)

	logger.Log(
		logger.Debug,
		"loaded command [%s] data %+v",
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
	return m.cmdTree
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
	logger.Log(logger.Debug, "stored cmd map [%+v] with [%s]", m.cmdMap, cmdPath)
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

func validateBranches[T any](cmds []BranchCommand[T], cmdTree Tree) {
	branchMap := map[string]struct{}{}
	for _, branch := range cmdTree.Branches {
		cmdPath := strings.Join(branch.Leaves, " ")
		if cmdPath == "" && branch.NeedArg && len(cmdTree.Branches) > 1 {
			logger.LogFatal(
				"[%s] is using itself as a default command with both args and branch commands",
				MsgIsCmdItself,
				cmdTree.Name,
			)
		}

		if branch.NeedArg && branch.ValidateArg == nil {
			logger.LogFatal(
				"Branch [%s] is missing an arg validation function.",
				MsgMissingArgFuncErr,
				cmdPath,
			)
		}
		if _, ok := branchMap[cmdPath]; ok {
			logger.LogFatal("Found duplicate tree branches [%s]", "", cmdPath)
		}

		foundCmd := slices.ContainsFunc(cmds, func(cmd BranchCommand[T]) bool {
			return cmd.Path == cmdPath
		})
		if !foundCmd {
			panic("branch command is missing a command branch")
		}

		branchMap[cmdPath] = struct{}{}
	}
}

func mapCommands[T any](cmds []BranchCommand[T], cmdTree Tree) map[string]BranchCommand[T] {
	cmdMap := map[string]BranchCommand[T]{}
	for _, cmd := range cmds {
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
