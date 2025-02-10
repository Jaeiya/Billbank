package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib/ui"
	"github.com/jaeiya/billbank/lib/utils/logger"
)

type (
	ExecBranchMsg struct {
		isOnKey          bool
		isOnViewportSize bool
	}
	ViewportSizeMsg struct {
		Width  int
		Height int
	}
)

type Branch[T any] struct {
	String       string
	Fn           func(T) T
	ViewFn       func(T) string
	IsPollingKey bool
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
		logger.Log(logger.Debug, fmt.Sprintf("BaseModel: setting viewport size [%dx%d]", msg.Width, msg.Height))
		bc.viewHeight = msg.Height
		bc.viewWidth = msg.Width
		// We don't need to execute the branch command on first
		// msg because it will be sent by the viewport.
		if !bc.isFirstMsg {
			teaCmds = append(teaCmds, func() tea.Msg { return ExecBranchMsg{isOnViewportSize: true} })
		}

	case Status:
		bc.cmdStatus = msg
		teaCmds = append(teaCmds, func() tea.Msg { return ExecBranchMsg{} })

	case tea.KeyMsg:
		teaCmds = append(teaCmds, func() tea.Msg { return ExecBranchMsg{isOnKey: true} })

	case ExecBranchMsg:
		cmd := bc.cmdBranchMap[bc.cmdStatus.BranchStr]
		if msg.isOnKey && cmd.IsPollingKey || msg.isOnViewportSize {
			model = bc.exec(model, false)
		} else if !msg.isOnKey && !msg.isOnViewportSize {
			model = bc.exec(model, true)
		}

	}

	return model, tea.Batch(teaCmds...)
}

func (bc *BaseCmdModel[T]) View(model T) string {
	branchStr := bc.cmdStatus.BranchStr
	cmd := bc.cmdBranchMap[branchStr]

	if cmd.ViewFn == nil {
		bc.AddError(
			fmt.Errorf(
				"tried to display missing view from [%s]",
				branchStr,
			),
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

	return cmd.ViewFn(model)
}

func (bc BaseCmdModel[T]) GetViewSize() (int, int) {
	return bc.viewWidth, bc.viewHeight
}

func (bc *BaseCmdModel[T]) AddBranch(branchCmds ...Branch[T]) {
	for _, cmd := range branchCmds {
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
				fmt.Sprintf(
					"CommandError: [%s] is missing an implementation func()",
					cmd.String,
				),
			)
		}
		if cmd.ViewFn == nil {
			logger.Log(
				logger.Error,
				fmt.Sprintf("CommandError: [%s] is missing a view func()", cmd.String),
			)
		}
	}
}

func (bc BaseCmdModel[T]) IsSupported(branchStr string) bool {
	_, ok := bc.cmdBranchMap[branchStr]
	return ok
}

// IsInitialized checks to make sure that the command not
// only has available commands, but also that a status
// has been set.
func (bc BaseCmdModel[T]) IsInitialized() bool {
	return len(bc.cmdBranchMap) > 0 && len(bc.cmdStatus.BranchStr) > 0
}

// exec executes the current branch command in the context of the
// passed model, with the option to clear all past and present
// errors. All detected errors are logged and stored.
func (bc *BaseCmdModel[T]) exec(model T, clearErrors bool) T {
	branchStr := bc.cmdStatus.BranchStr
	cmd := bc.cmdBranchMap[branchStr]

	if clearErrors {
		bc.ClearErrors()
	}

	withOrWithoutErr := "with error"
	if bc.lastCmdError.Error() == "" {
		withOrWithoutErr = "without error"
	}

	logger.Log(
		logger.Debug,
		fmt.Sprintf(
			"BaseModel: executing branch [%s] [%s]",
			branchStr,
			withOrWithoutErr,
		),
	)

	if cmd.Fn == nil {
		err := fmt.Errorf("[%s] tried to execute missing implementation func()", branchStr)
		if bc.lastCmdError.Error() != err.Error() {
			logger.Log(logger.Error, fmt.Sprintf("CommandError: %s", err))
			bc.lastCmdError = err
			bc.AddError(err)
		}
		return model
	}

	if cmd.ViewFn == nil {
		err := fmt.Errorf("[%s] tried to execute missing view func()", branchStr)
		if bc.lastCmdError.Error() != err.Error() {
			logger.Log(logger.Error, fmt.Sprintf("CommandError: %s", err))
			bc.lastCmdError = err
			bc.AddError(err)
		}
		return model
	}

	model = cmd.Fn(model)
	errs := bc.GetErrors()
	if len(errs) > 0 {
		if bc.lastCmdError.Error() != errs[0].Error() {
			logger.Log(logger.Error, fmt.Sprintf("CommandError: %s", errs[0].Error()))
		}
		bc.lastCmdError = errs[0]
	}

	return model
}
