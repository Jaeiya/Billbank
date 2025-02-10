package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib/ui"
	"github.com/jaeiya/billbank/lib/utils/logger"
)

type (
	ViewportSizeMsg struct {
		Width  int
		Height int
	}
)

type Branch[T any] struct {
	String string
	Fn     func(T) T
	ViewFn func(T) string
}

type BaseCmdModel[T any] struct {
	cmdBranchMap map[string]Branch[T]
	cmdTree      [][]string
	cmdStatus    Status
	cmdErrors    []error
	lastCmdError error
	isFirstMsg   bool
	viewWidth    int
	viewHeight   int
	hasStaleView bool
	staleView    string
}

func NewBaseModel[T any](tree [][]string) *BaseCmdModel[T] {
	return &BaseCmdModel[T]{
		cmdBranchMap: map[string]Branch[T]{},
		cmdTree:      tree,
		lastCmdError: fmt.Errorf(""),
		isFirstMsg:   true,
	}
}

func (bc *BaseCmdModel[T]) Update(model T, msg tea.Msg) (T, tea.Cmd) {
	var teaCmds []tea.Cmd
	defer func() { bc.isFirstMsg = false }()

	switch msg := msg.(type) {
	case ViewportSizeMsg:
		logger.Log(logger.Hot, "BaseModel", " setting viewport size [%dx%d]", msg.Width, msg.Height)
		bc.viewHeight = msg.Height
		bc.viewWidth = msg.Width
		bc.hasStaleView = false
		model = bc.Exec(model)

	case UpdateCmdMsg:
		// The branch won't be executed until the next model update
		// therefore we need to mark the view as stale, so it won't
		// try to view uninitialized model data.
		oldBranch := bc.cmdStatus.BranchStr
		bc.hasStaleView = true
		bc.cmdStatus = msg.Status
		logger.Log(logger.Debug, "BaseModel", " updated command branch [%s] to [%s]", oldBranch, bc.cmdStatus.BranchStr)

	}

	return model, tea.Batch(teaCmds...)
}

func (bc *BaseCmdModel[T]) View(model T) string {
	branchStr := bc.cmdStatus.BranchStr
	cmd := bc.cmdBranchMap[branchStr]

	if bc.hasStaleView {
		logger.Log(logger.Debug, "BaseModel", " loading stale view [%s]", branchStr)
		return bc.staleView
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

	bc.staleView = cmd.ViewFn(model)
	return cmd.ViewFn(model)
}

func (bc BaseCmdModel[T]) GetViewSize() (int, int) {
	return bc.viewWidth, bc.viewHeight
}

func (bc *BaseCmdModel[T]) AddBranch(branchCmds ...Branch[T]) {
	for _, cmd := range branchCmds {
		if _, alreadyExists := bc.cmdBranchMap[cmd.String]; alreadyExists {
			panic(fmt.Errorf("found multiple command branches for [%s]", cmd.String))
		}
		bc.cmdBranchMap[cmd.String] = cmd
	}
}

func (bc *BaseCmdModel[T]) AddError(err error) {
	bc.cmdErrors = append(bc.cmdErrors, err)
}

func (bc BaseCmdModel[T]) GetErrors() []error {
	return bc.cmdErrors
}

func (bc *BaseCmdModel[T]) ClearErrors() {
	if len(bc.cmdErrors) > 0 {
		bc.lastCmdError = fmt.Errorf("")
		bc.cmdErrors = nil
	}
}

func (bc BaseCmdModel[T]) GetCmdTree() [][]string {
	return bc.cmdTree
}

func (bc BaseCmdModel[T]) IsActiveBranch(branch string) bool {
	return bc.cmdStatus.BranchStr == branch
}

/*
HasView returns true if the current command tree string has
an applicable view associated with it.
*/
func (bc BaseCmdModel[T]) HasView() bool {
	cmd := bc.cmdBranchMap[bc.cmdStatus.BranchStr]
	return cmd.ViewFn != nil
}

func (bc BaseCmdModel[T]) ValidateCommand() {
	if len(bc.cmdBranchMap) == 0 {
		panic("missing sub commands, did you forget to add them?")
	}

	for _, cmd := range bc.cmdBranchMap {
		if cmd.Fn == nil {
			logger.Log(
				logger.Error,
				"CommandError",
				"[%s] is missing an implementation func()",
				cmd.String,
			)
		}
		if cmd.ViewFn == nil {
			logger.Log(
				logger.Error,
				"CommandError",
				"[%s] is missing a view func()",
				cmd.String,
			)
		}
	}
}

func (bc BaseCmdModel[T]) IsSupported(branchStr string) bool {
	_, ok := bc.cmdBranchMap[branchStr]
	return ok
}

// IsInitialized checks to make sure that various expected values
// are set.
func (bc BaseCmdModel[T]) IsInitialized() bool {
	return len(bc.cmdBranchMap) > 0 && len(bc.cmdStatus.BranchStr) > 0 && bc.viewWidth > 0 &&
		bc.viewHeight > 0
}

// Exec executes the current branch command in the context of the
// passed model, with the option to clear all past and present
// errors. All detected errors are logged and stored.
func (bc *BaseCmdModel[T]) Exec(model T) T {
	branchStr := bc.cmdStatus.BranchStr
	cmd := bc.cmdBranchMap[branchStr]

	bc.ClearErrors()

	logger.Log(logger.Hot, "BaseModel", " executing branch [%s]", branchStr)

	if cmd.Fn == nil {
		err := fmt.Errorf("[%s] tried to execute missing implementation func()", branchStr)
		if bc.lastCmdError.Error() != err.Error() {
			logger.Log(logger.Error, "CommandError", "%s", err)
			bc.lastCmdError = err
			bc.AddError(err)
		}
		return model
	}

	if cmd.ViewFn == nil {
		err := fmt.Errorf("[%s] tried to execute missing view func()", branchStr)
		if bc.lastCmdError.Error() != err.Error() {
			logger.Log(logger.Error, "CommandError", "%s", err)
			bc.lastCmdError = err
			bc.AddError(err)
		}
		return model
	}

	model = cmd.Fn(model)
	errs := bc.GetErrors()
	if len(errs) > 0 {
		if bc.lastCmdError.Error() != errs[0].Error() {
			logger.Log(logger.Error, "CommandError", "%s", errs[0].Error())
		}
		bc.lastCmdError = errs[0]
	}

	return model
}
