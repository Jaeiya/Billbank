package commander

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
	CmdViewportSizeMsg struct {
		Width  int
		Height int
	}
)

type BranchFunc[T any] func(T) T

type BranchCommand[T any] struct {
	String       string
	Fn           BranchFunc[T]
	ViewFn       func(T) string
	isPollingKey bool
}

type BaseCommand[T any] struct {
	cmdMap     map[string]BranchCommand[T]
	cmdTree    [][]string
	cmdStatus  CommandStatus
	cmdErrors  []error
	lastError  error
	viewWidth  int
	viewHeight int
}

func NewBaseCommand[T any](tree [][]string) *BaseCommand[T] {
	return &BaseCommand[T]{
		cmdMap:    map[string]BranchCommand[T]{},
		cmdTree:   tree,
		lastError: fmt.Errorf(""),
	}
}

func (bc *BaseCommand[T]) Update(model T, msg tea.Msg) (T, tea.Cmd) {
	var teaCmds []tea.Cmd

	switch msg := msg.(type) {
	case CmdViewportSizeMsg:
		logger.Log(logger.Debug, fmt.Sprintf("BaseCommand: setting viewport size [%dx%d]", msg.Width, msg.Height))
		bc.viewHeight = msg.Height
		bc.viewWidth = msg.Width
		if bc.cmdStatus.BranchStr != "" {
			teaCmds = append(teaCmds, func() tea.Msg { return ExecBranchMsg{isOnViewportSize: true} })
		}

	case CommandStatus:
		bc.cmdStatus = msg
		teaCmds = append(teaCmds, func() tea.Msg { return ExecBranchMsg{} })

	case tea.KeyMsg:
		teaCmds = append(teaCmds, func() tea.Msg { return ExecBranchMsg{isOnKey: true} })

	case ExecBranchMsg:
		cmd := bc.cmdMap[bc.cmdStatus.BranchStr]
		if msg.isOnKey && cmd.isPollingKey || msg.isOnViewportSize {
			model = bc.exec(model, false)
		} else if !msg.isOnKey && !msg.isOnViewportSize {
			model = bc.exec(model, true)
		}

	}

	return model, tea.Batch(teaCmds...)
}

func (bc *BaseCommand[T]) View(model T) string {
	branchStr := bc.cmdStatus.BranchStr
	cmd := bc.cmdMap[branchStr]

	if cmd.ViewFn == nil {
		bc.AddError(
			fmt.Errorf(
				"tried to display missing view from [%s]",
				bc.cmdStatus.BranchStr,
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

func (bc *BaseCommand[T]) AddBranch(branchCmds ...BranchCommand[T]) {
	for _, cmd := range branchCmds {
		bc.cmdMap[cmd.String] = cmd
	}
}

func (bc *BaseCommand[T]) AddError(err error) {
	bc.cmdErrors = append(bc.cmdErrors, err)
}

func (bc BaseCommand[T]) GetErrors() []error {
	return bc.cmdErrors
}

func (bc *BaseCommand[T]) ClearErrors() {
	if len(bc.cmdErrors) > 0 {
		bc.lastError = fmt.Errorf("")
		bc.cmdErrors = nil
	}
}

func (bc BaseCommand[T]) GetCmdTree() [][]string {
	return bc.cmdTree
}

/*
HasView returns true if the current command tree string has
an applicable view associated with it.
*/
func (bc BaseCommand[T]) HasView() bool {
	cmd := bc.cmdMap[bc.cmdStatus.BranchStr]
	return cmd.ViewFn != nil
}

func (bc BaseCommand[T]) ValidateCommand() {
	if len(bc.cmdMap) == 0 {
		panic("missing sub commands, did you forget to add them?")
	}

	for _, cmd := range bc.cmdMap {
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

func (bc BaseCommand[T]) IsSupported(branchStr string) bool {
	_, ok := bc.cmdMap[branchStr]
	return ok
}

// IsInitialized checks to make sure that the command not
// only has available commands, but also that a status
// has been set.
func (bc BaseCommand[T]) IsInitialized() bool {
	return len(bc.cmdMap) > 0 && len(bc.cmdStatus.BranchStr) > 0
}

func (bc *BaseCommand[T]) exec(model T, clearErrors bool) T {
	branchStr := bc.cmdStatus.BranchStr
	cmd := bc.cmdMap[branchStr]

	if clearErrors {
		bc.ClearErrors()
	}

	if cmd.Fn == nil {
		err := fmt.Errorf("[%s] tried to execute missing implementation func()", branchStr)
		if bc.lastError.Error() != err.Error() {
			logger.Log(logger.Error, fmt.Sprintf("CommandError: %s", err))
			bc.lastError = err
			bc.AddError(err)
		}
		return model
	}

	if cmd.ViewFn == nil {
		err := fmt.Errorf("[%s] tried to execute missing view func()", branchStr)
		if bc.lastError.Error() != err.Error() {
			logger.Log(logger.Error, fmt.Sprintf("CommandError: %s", err))
			bc.lastError = err
			bc.AddError(err)
		}
		return model
	}

	return cmd.Fn(model)
}
