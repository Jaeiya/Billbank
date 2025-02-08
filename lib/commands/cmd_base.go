package commands

import (
	"fmt"
	"time"
	"unsafe"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib/ui"
	"github.com/jaeiya/billbank/lib/utils/logger"
)

type ExecCmdMsg string

type CommandFunc[T any] func(*T) tea.Cmd

type CommandEntry[T any] struct {
	String string
	Fn     CommandFunc[T]
	ViewFn func(T) string
}

type BaseCommand[T any] struct {
	cmdList    []string
	cmdMap     map[string]CommandFunc[T]
	cmdViewMap map[string]func(T) string
	cmdTree    [][]string
	cmdStatus  ui.CommandStatus
	cmdError   error
	viewWidth  int
	viewHeight int
}

func NewBaseCommand[T any](tree [][]string) BaseCommand[T] {
	return BaseCommand[T]{
		cmdMap:     map[string]CommandFunc[T]{},
		cmdViewMap: map[string]func(T) string{},
		cmdTree:    tree,
	}
}

func (bc *BaseCommand[T]) Update(model *T, msg tea.Msg) (*T, tea.Cmd) {
	basePtr := (*BaseCommand[T])(unsafe.Pointer(model))

	switch msg := msg.(type) {
	case ui.ViewportSizeMsg:
		logger.Log(logger.Debug, fmt.Sprintf("ViewPortSize: %dx%d", msg.Width, msg.Height))
		// Hack to get around type safety
		basePtr.viewHeight = msg.Height
		basePtr.viewWidth = msg.Width
		return model, func() tea.Msg { return ExecCmdMsg(bc.cmdStatus.TreeStr) }

	case ExecCmdMsg:
		// Do not propagate errors to new commands
		if basePtr.cmdError != nil {
			basePtr.SetError(nil)
		}
		cmd, err := bc.Exec(model)
		if err != nil {
			basePtr.SetError(err)
		} else if !bc.HasView() {
			basePtr.SetError(fmt.Errorf("tried to display missing view from [%s]", bc.cmdStatus.TreeStr))
		}

		return model, cmd

	}

	return model, nil
}

func (bc *BaseCommand[T]) AddCommands(cmds ...CommandEntry[T]) {
	for _, cmd := range cmds {
		bc.cmdList = append(bc.cmdList, cmd.String)
		if cmd.Fn != nil {
			bc.cmdMap[cmd.String] = cmd.Fn
		}
		if cmd.ViewFn != nil {
			bc.cmdViewMap[cmd.String] = cmd.ViewFn
		}
	}
}

func (bc *BaseCommand[T]) SetStatus(status ui.CommandStatus) tea.Cmd {
	bc.cmdStatus = status
	return func() tea.Msg { return ExecCmdMsg(bc.cmdStatus.TreeStr) }
}

/*
SetError Assigns the specified error to the current commands
error field.

🟠 Can only be used in a function where the parent model is
being returned, otherwise it will do nothing.
*/
func (bc *BaseCommand[T]) SetError(err error) {
	bc.cmdError = err
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
	_, ok := bc.cmdViewMap[bc.cmdStatus.TreeStr]
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

func (bc BaseCommand[T]) IsSupported(treeStr string) bool {
	_, ok := bc.cmdMap[treeStr]
	return ok
}

func (bc BaseCommand[T]) Exec(model *T) (tea.Cmd, error) {
	cmd := bc.cmdStatus.TreeStr
	fn, ok := bc.cmdMap[cmd]
	if !ok {
		return nil, fmt.Errorf("command::[%s] not implemented", cmd)
	}
	return fn(model), nil
}

func (bc BaseCommand[T]) ExecView(model T) string {
	cmd := bc.cmdStatus.TreeStr
	fn, ok := bc.cmdViewMap[cmd]
	if !ok {
		return fmt.Sprintf("command::[%s] missing view", cmd)
	}
	return fn(model)
}

func (bc BaseCommand[T]) poll(d time.Duration) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(d)
		return ExecCmdMsg(bc.cmdStatus.TreeStr)
	}
}
