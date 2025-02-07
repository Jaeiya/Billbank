package commands

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib/ui"
	"github.com/jaeiya/billbank/lib/utils/logger"
)

type CommandEntry[T any] struct {
	String string
	Fn     func(T) T
	ViewFn func(T) string
}

type BaseCommand[T any] struct {
	cmdList    []string
	cmdMap     map[string]func(T) T
	cmdViewMap map[string]func(T) string
	cmdTree    [][]string
	cmdStatus  ui.CommandStatus
	cmdError   error
	viewWidth  int
	viewHeight int
}

func NewBaseCommand[T any](tree [][]string) BaseCommand[T] {
	return BaseCommand[T]{
		cmdMap:     map[string]func(T) T{},
		cmdViewMap: map[string]func(T) string{},
		cmdTree:    tree,
	}
}

func (bc *BaseCommand[T]) Update(msg tea.Msg) {
	switch msg := msg.(type) {
	case ui.ViewportSizeMsg:
		logger.Log(logger.Debug, fmt.Sprintf("ViewPortSize: %dx%d", msg.Width, msg.Height))
		bc.viewHeight = msg.Height
		bc.viewWidth = msg.Width
	}
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

func (bc *BaseCommand[T]) SetStatus(status ui.CommandStatus) {
	bc.cmdStatus = status
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

func (bc BaseCommand[T]) Exec(model T) (T, error) {
	cmd := bc.cmdStatus.TreeStr
	fn, ok := bc.cmdMap[cmd]
	if !ok {
		return model, fmt.Errorf("command::[%s] not implemented", cmd)
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
