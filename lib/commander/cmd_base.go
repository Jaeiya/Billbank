package commander

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib/utils/logger"
)

type (
	ExecBranchMsg      string
	CmdViewportSizeMsg struct {
		Width  int
		Height int
	}
)

type BranchFunc[T any] func(*T) tea.Cmd

type BranchEntry[T any] struct {
	String string
	Fn     BranchFunc[T]
	ViewFn func(T) string
}

type BaseCommand[T any] struct {
	cmdList    []string
	cmdMap     map[string]BranchFunc[T]
	cmdViewMap map[string]func(T) string
	cmdTree    [][]string
	cmdStatus  CommandStatus
	cmdError   error
	viewWidth  int
	viewHeight int
}

func NewBaseCommand[T any](tree [][]string) *BaseCommand[T] {
	return &BaseCommand[T]{
		cmdMap:     map[string]BranchFunc[T]{},
		cmdViewMap: map[string]func(T) string{},
		cmdTree:    tree,
	}
}

func (bc *BaseCommand[T]) Update(model *T, msg tea.Msg) (*T, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case CmdViewportSizeMsg:
		logger.Log(logger.Debug, fmt.Sprintf("BaseCommand: setting viewport size [%dx%d]", msg.Width, msg.Height))
		bc.viewHeight = msg.Height
		bc.viewWidth = msg.Width
		return model, func() tea.Msg { return ExecBranchMsg(bc.cmdStatus.BranchStr) }

	case CommandStatus:
		// Do not re-execute branch command if it's already running
		if bc.cmdStatus.BranchStr != msg.BranchStr {
			bc.cmdStatus = msg
			logger.Log(logger.Debug, fmt.Sprintf("BaseCommand: executing branch [%s]", msg.BranchStr))
			cmds = append(cmds, func() tea.Msg { return ExecBranchMsg(msg.BranchStr) })
		}

	case ExecBranchMsg:
		// Each execution is considered a new command execution
		// therefore we treat it as a "first" execution.
		if bc.cmdError != nil {
			bc.cmdError = nil
		}
		cmds = append(cmds, bc.exec(model))

	}

	return model, tea.Batch(cmds...)
}

func (bc BaseCommand[T]) View(model T) string {
	branchStr := bc.cmdStatus.BranchStr
	fn, ok := bc.cmdViewMap[bc.cmdStatus.BranchStr]
	if !ok {
		return fmt.Sprintf("command::[%s] missing view", branchStr)
	}
	return fn(model)
}

func (bc *BaseCommand[T]) AddBranch(branches ...BranchEntry[T]) {
	for _, branch := range branches {
		bc.cmdList = append(bc.cmdList, branch.String)
		if branch.Fn != nil {
			bc.cmdMap[branch.String] = branch.Fn
		}
		if branch.ViewFn != nil {
			bc.cmdViewMap[branch.String] = branch.ViewFn
		}
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
	_, ok := bc.cmdViewMap[bc.cmdStatus.BranchStr]
	return ok
}

func (bc BaseCommand[T]) ValidateCommand() {
	if len(bc.cmdMap) == 0 {
		panic("missing sub commands, did you forget to add them?")
	}

	for _, cmd := range bc.cmdList {
		if _, ok := bc.cmdMap[cmd]; !ok {
			logger.Log(
				logger.Error,
				fmt.Sprintf("CommandError: command not implemented for [%s]", cmd),
			)
		}
	}

	for cmd := range bc.cmdMap {
		_, ok := bc.cmdViewMap[cmd]
		if !ok {
			logger.Log(
				logger.Attention,
				fmt.Sprintf("CommandWarn: missing command view for [%s]", cmd),
			)
		}
	}
}

func (bc BaseCommand[T]) IsSupported(branchStr string) bool {
	_, ok := bc.cmdMap[branchStr]
	return ok
}

func (bc *BaseCommand[T]) exec(model *T) tea.Cmd {
	branchStr := bc.cmdStatus.BranchStr
	fn, ok := bc.cmdMap[branchStr]
	if !ok {
		bc.cmdError = fmt.Errorf("command::[%s] not implemented", branchStr)
	} else if !bc.HasView() {
		bc.cmdError = fmt.Errorf("tried to display missing view from [%s]", bc.cmdStatus.BranchStr)
	}
	return fn(model)
}

func (bc BaseCommand[T]) poll(d time.Duration) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(d)
		return ExecBranchMsg(bc.cmdStatus.BranchStr)
	}
}
