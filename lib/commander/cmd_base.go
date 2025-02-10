package commander

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib/utils/logger"
)

type (
	ExecBranchMsg      struct{}
	ExecOnKey          struct{}
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
	isKeyPolling bool
}

type BaseCommand[T any] struct {
	cmdMap     map[string]BranchCommand[T]
	cmdTree    [][]string
	cmdStatus  CommandStatus
	cmdError   error
	viewWidth  int
	viewHeight int
}

func NewBaseCommand[T any](tree [][]string) *BaseCommand[T] {
	return &BaseCommand[T]{
		cmdMap:  map[string]BranchCommand[T]{},
		cmdTree: tree,
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
			teaCmds = append(teaCmds, func() tea.Msg { return ExecBranchMsg{} })
		}

	case CommandStatus:
		bc.cmdStatus = msg
		teaCmds = append(teaCmds, func() tea.Msg { return ExecBranchMsg{} })

	case tea.KeyMsg:
		teaCmds = append(teaCmds, func() tea.Msg { return ExecOnKey{} })

	case ExecOnKey:
		cmd := bc.cmdMap[bc.cmdStatus.BranchStr]
		if cmd.isKeyPolling {
			model = bc.exec(model)
		}

	case ExecBranchMsg:
		model = bc.exec(model)

	}

	return model, tea.Batch(teaCmds...)
}

func (bc *BaseCommand[T]) View(model T) string {
	branchStr := bc.cmdStatus.BranchStr
	cmd := bc.cmdMap[branchStr]

	if cmd.ViewFn == nil {
		bc.cmdError = fmt.Errorf(
			"tried to display missing view from [%s]",
			bc.cmdStatus.BranchStr,
		)
		return fmt.Sprintf("command::[%s] missing view", branchStr)
	}

	return cmd.ViewFn(model)
}

func (bc *BaseCommand[T]) AddBranch(branchCmds ...BranchCommand[T]) {
	for _, cmd := range branchCmds {
		bc.cmdMap[cmd.String] = cmd
	}
}

func (bc BaseCommand[T]) GetError() error {
	return bc.cmdError
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
				fmt.Sprintf("CommandError: command not implemented for [%s]", cmd.String),
			)
		}
		if cmd.ViewFn == nil {
			logger.Log(
				logger.Attention,
				fmt.Sprintf("CommandWarn: missing command view for [%s]", cmd.String),
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

func (bc *BaseCommand[T]) exec(model T) T {
	branchStr := bc.cmdStatus.BranchStr
	cmd := bc.cmdMap[branchStr]

	if bc.cmdError != nil {
		bc.cmdError = nil
	}

	if cmd.Fn == nil {
		bc.cmdError = fmt.Errorf("command::[%s] not implemented", branchStr)
	}

	if cmd.ViewFn == nil {
		bc.cmdError = fmt.Errorf(
			"tried to display missing view from [%s]",
			bc.cmdStatus.BranchStr,
		)
	}
	return cmd.Fn(model)
}
