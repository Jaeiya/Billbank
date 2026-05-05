package cmdcore

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/jaeiya/billbank/internal/logger"
)

type ArgType int

const (
	ArgNone = ArgType(iota)
	ArgOptional
	ArgRequired
)

var (
	ErrModelTypeMismatch       = fmt.Errorf("model type mismatch")
	ErrModelTypeReturnMismatch = fmt.Errorf("command has returned the wrong model type")
	ErrNoResolveArg            = fmt.Errorf("command does not have an argument to resolve")
)

type Command[M any] interface {
	GetPath() string
	GetArgType() ArgType
	CanCaptureInput() bool

	// SetWorkingPath will be used to let the command model know
	// which command path is active in the command handler.
	SetWorkingPath(path string)

	// Validate that the model passed is M otherwise error
	Run(model any) (CommandModel, error)

	// Validate that the model passed is M otherwise error
	View(model any) (tea.View, error)

	// Validates an argument passed to the command & should
	// error if args are not supported.
	ResolveArg(arg string) error
}

type runFunc[M any, A any] func(model M, arg *A) (CommandModel, error)

type CommandOptions[M any, A any] struct {
	// A word or string of words separated by a space, which lead to
	// the execution of the command.
	//
	// Ex: "clear" or "clear log" or "clear history"
	//
	// 🟡 An empty path refers to the command alias itself as the path.
	Path string

	// Hides the command input, which relinquishes keyboard control to
	// the command. This is necessary for commands which control the
	// UI using the keyboard.
	CaptureInput bool

	// The type of arguments that the command requires.
	// ArgNone is the default.
	ArgType ArgType

	// Executes a commands logic and should return the passed model.
	//
	// 🟠 An error should be returned ONLY if the error can be
	// considered an unrecoverable event; all other errors
	// should be handled by the commands Run and View
	// methods.
	RunFunc runFunc[M, A]

	// Executes the view for the command
	ViewFunc func(model M) tea.View

	ParseFunc func(arg string) (A, error)
}

type command[M any, A any] struct {
	Path         string
	CaptureInput bool
	ArgType      ArgType
	RunFunc      runFunc[M, A]
	ViewFunc     func(model M) tea.View
	ParseFunc    func(arg string) (A, error)
	arg          *A
	workingPath  string
}

func (c command[M, A]) GetPath() string {
	return c.Path
}

func (c command[M, A]) GetArgType() ArgType {
	return c.ArgType
}

func (c command[M, A]) CanCaptureInput() bool {
	return c.CaptureInput
}

func (c *command[M, A]) SetWorkingPath(s string) {
	c.workingPath = s
}

func (c command[M, A]) IsWorkingPath(path string) bool {
	return c.workingPath == path
}

func (c *command[M, A]) Run(m any) (CommandModel, error) {
	if v, isType := m.(M); isType {
		var argCopy *A
		if c.arg != nil {
			// Arguments should always be consumed; never remembered
			argCopy = new(*c.arg)
			c.arg = nil
		}
		rm, err := c.RunFunc(v, argCopy)
		if _, isType = rm.(M); !isType {
			return nil, ErrModelTypeReturnMismatch
		}
		if err != nil {
			return rm, err
		}
		return rm, nil
	}
	return nil, ErrModelTypeMismatch
}

func (c command[M, A]) View(m any) (tea.View, error) {
	if v, isType := m.(M); isType {
		return c.ViewFunc(v), nil
	}
	return tea.NewView(""), ErrModelTypeMismatch
}

func (c *command[M, A]) ResolveArg(arg string) error {
	if len(arg) == 0 {
		if c.ArgType == ArgNone {
			return ErrNoResolveArg
		}
		if c.ArgType == ArgOptional {
			return nil
		}
	}

	parsedArg, err := c.ParseFunc(arg)
	if err != nil {
		return err
	}

	c.arg = &parsedArg
	return nil
}

// type noArgCommand[M any] struct {
// 	command[M, A]
// }

// func NewNoArgCommand[M any](opt CommandOptions[M]) Command[M] {
// 	return noArgCommand[M]{
// 		command: command[M]{
// 			Path:         opt.Path,
// 			CaptureInput: opt.CaptureInput,
// 			RunFunc:      opt.RunFunc,
// 			ViewFunc:     opt.ViewFunc,
// 		},
// 	}
// }

// func (c noArgCommand[M]) ResolveArg(arg string) error {
// 	return ErrNoResolveArg
// }

type ArgParser[T any] func(arg string) (T, error)

type NoArg struct{}

func NewCommand[M any, A any](opt CommandOptions[M, A]) Command[M] {
	if opt.RunFunc == nil {
		logger.LogFatal(
			"command [%s] is missing a RunFunc",
			"Did you forget to assign the RunFunc field for this command?",
			opt.Path,
		)
	}

	if opt.ViewFunc == nil {
		logger.LogFatal(
			"command [%s] is missing a ViewFunc",
			"Did you forget to assign the ViewFunc field for this command?",
			opt.Path,
		)
	}

	if opt.ArgType == ArgNone && opt.ParseFunc != nil {
		logger.LogFatal(
			"command [%s] expects an arg type to be parsed",
			"Did you forget to set the arg type for this command?",
			opt.Path,
		)
	}

	if opt.ArgType > ArgNone && opt.ParseFunc == nil {
		logger.LogFatal(
			"command [%s] is missing ParseFunc",
			"This command is setup to accept args, did you forget?",
			opt.Path,
		)
	}

	return &command[M, A]{
		Path:         opt.Path,
		ArgType:      opt.ArgType,
		CaptureInput: opt.CaptureInput,
		RunFunc:      opt.RunFunc,
		ViewFunc:     opt.ViewFunc,
		ParseFunc:    opt.ParseFunc,
	}
}

// type argCommand[M any, A any] struct {
// 	command[M]
// 	lastArg   *A
// 	parseFunc ArgParser[A]
// }

// func NewArgCommand[M any, A any](opt CommandOptions[M], p ArgParser[A]) *argCommand[M, A] {
// 	if opt.ArgType == ArgNone {
// 		logger.LogFatal(
// 			"cannot use ArgNone with command [%s]",
// 			"You either forgot to set the arg type for this command or you used the wrong command constructor.",
// 			opt.Path,
// 		)
// 	}

// 	return &argCommand[M, A]{
// 		command: command[M]{
// 			Path:         opt.Path,
// 			ArgType:      opt.ArgType,
// 			CaptureInput: opt.CaptureInput,
// 			RunFunc:      opt.RunFunc,
// 			ViewFunc:     opt.ViewFunc,
// 		},
// 		parseFunc: p,
// 	}
// }

// func (c *argCommand[M, A]) ResolveArg(arg string) error {
// 	// Prevents potential undefined behavior if somehow
// 	// the lastArg is referenced after this func is called
// 	c.lastArg = nil

// 	v, err := c.parseFunc(arg)
// 	if err != nil {
// 		return err
// 	}
// 	c.lastArg = &v
// 	return nil
// }

// func (c argCommand[M, A]) GetArg() A {
// 	return *c.lastArg
// }
