package cmd

import (
	"fmt"

	"github.com/charmbracelet/bubbles/cursor"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib/logger"
	"github.com/jaeiya/billbank/lib/ui"
)

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
	cmdMap := map[string]BranchCommand[T]{}
	for _, cmd := range cmds {
		if _, alreadyExists := cmdMap[cmd.Path]; alreadyExists {
			panic(fmt.Errorf("found multiple command paths for [%s]", cmd.Path))
		}
		cmdMap[cmd.Path] = cmd
	}

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
		logger.Log(
			logger.Hot, "BaseModel", "setting viewport size [%dx%d]", msg.Width, msg.Height,
		)
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
		logger.Log(
			logger.Debug,
			"BaseModel",
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
		logger.Log(logger.Hot, "BaseModel", "loading [stale] view [%s]", bc.stalePath)
		return bc.staleView
	}

	if !bc.isInterrupt {
		logger.Log(
			logger.Hot,
			"BaseModel",
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

func (bc BaseModel[T]) GetViewSize() (int, int) {
	return bc.viewWidth, bc.viewHeight
}

func (bc *BaseModel[T]) AddError(err error) {
	bc.cmdErrors = append(bc.cmdErrors, err)
}

func (bc BaseModel[T]) GetErrors() []error {
	return bc.cmdErrors
}

func (bc *BaseModel[T]) ClearErrors() {
	if len(bc.cmdErrors) > 0 {
		bc.lastCmdError = fmt.Errorf("")
		bc.cmdErrors = nil
	}
}

func (bc BaseModel[T]) GetCmdTree() Tree {
	return bc.cmdTree
}

func (bc BaseModel[T]) IsActivePath(cmdPath string) bool {
	return bc.cmdStatus.Path == cmdPath
}

/*
HasView returns true if the current command tree string has
an applicable view associated with it.
*/
func (bc BaseModel[T]) HasView() bool {
	cmd := bc.cmdMap[bc.cmdStatus.Path]
	return cmd.View != nil
}

func (bc BaseModel[T]) ValidateCommand() {
	if len(bc.cmdMap) == 0 {
		panic("missing sub commands, did you forget to add them?")
	}

	for _, cmd := range bc.cmdMap {
		if cmd.Run == nil {
			logger.Log(
				logger.Error,
				"CommandError",
				"[%s] is missing an implementation func()",
				cmd.Path,
			)
		}
		if cmd.View == nil {
			logger.Log(
				logger.Error,
				"CommandError",
				"[%s] is missing a view func()",
				cmd.Path,
			)
		}
	}
}

func (bc BaseModel[T]) IsSupported(cmdPath string) bool {
	_, ok := bc.cmdMap[cmdPath]
	return ok
}

// IsInitialized checks to make sure that various expected values
// are set.
func (bc BaseModel[T]) IsInitialized() bool {
	return len(bc.cmdMap) > 0 && len(bc.cmdStatus.Path) > 0 && bc.viewWidth > 0 &&
		bc.viewHeight > 0
}

// Exec executes the current command path in the context of the
// passed model. All detected errors are logged and stored.
func (bc *BaseModel[T]) Exec(model T) T {
	bc.isInterrupt = false
	cmdPath := bc.cmdStatus.Path
	cmd := bc.cmdMap[cmdPath]

	bc.ClearErrors()

	logger.Log(logger.Hot, "BaseModel", "executing command path [%s]", cmdPath)

	if cmd.Run == nil {
		err := fmt.Errorf("[%s] tried to execute missing implementation func()", cmdPath)
		if bc.lastCmdError.Error() != err.Error() {
			logger.Log(logger.Error, "CommandError", "%s", err)
			bc.lastCmdError = err
			bc.AddError(err)
		}
		return model
	}

	if cmd.View == nil {
		err := fmt.Errorf("[%s] tried to execute missing view func()", cmdPath)
		if bc.lastCmdError.Error() != err.Error() {
			logger.Log(logger.Error, "CommandError", "%s", err)
			bc.lastCmdError = err
			bc.AddError(err)
		}
		return model
	}

	model = cmd.Run(model)
	errs := bc.GetErrors()
	if len(errs) > 0 {
		if bc.lastCmdError.Error() != errs[0].Error() {
			logger.Log(logger.Error, "CommandError", "%s", errs[0].Error())
		}
		bc.lastCmdError = errs[0]
	}

	return model
}
