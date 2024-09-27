package commands

import (
	"fmt"
	"strings"

	"github.com/jaeiya/billbank/lib"
)

type CommandError error

var (
	ErrFatalCommand = CommandError(
		fmt.Errorf("command parsing failed; this should not happen"),
	)
	ErrCommandNotFound = CommandError(fmt.Errorf("command not found"))
	ErrMissingArgument = CommandError(fmt.Errorf("missing argument"))
)

func getFatalCommandErr(cmd string) CommandError {
	return CommandError(fmt.Errorf("'%s' caused a fatal command parsing error", cmd))
}

type (
	CommandFunc func(args ...string)
)

type CommandResult struct {
	IsCommand   bool
	IsComplete  bool
	Suggestions []string
	Error       error
	Func        func()
}

type CommandConfig struct {
	CommandBase
}

type CommandBase struct {
	stages         [][]string
	validationFunc func(arg string) error
	hasArg         bool
	exec           CommandFunc
}

func (cb *CommandBase) ParseCommand(cmd string) CommandResult {
	cmdFields := strings.Fields(cmd)

	if len(cmdFields) > len(cb.stages) && !cb.hasArg {
		return CommandResult{Error: ErrCommandNotFound}
	}

	for stage, cmds := range cb.stages {
		if stage == len(cmdFields) || !isCommand(cmds, cmdFields[stage]) {
			var isCommand bool
			var suggestions []string
			if stage > 0 {
				isCommand = true
				suggestions = cb.stages[stage]
			}
			return CommandResult{
				IsCommand:   isCommand,
				Suggestions: suggestions,
				Error:       ErrCommandNotFound,
			}
		}

		if cb.hasArg {
			if len(cmdFields) == len(cb.stages) {
				return CommandResult{
					IsCommand: true,
					Error:     ErrMissingArgument,
				}
			}
			if stage+1 == len(cb.stages) {
				return CommandResult{
					IsCommand:  true,
					IsComplete: stage+1 == len(cb.stages),
					Func:       func() { cb.exec(cmdFields...) },
					Error:      cb.validationFunc(cmdFields[stage+1]),
				}
			}
		}

		if stage+1 < len(cmdFields) {
			continue
		}

		isComplete := stage+1 == len(cb.stages)
		var suggestions []string
		f := func() { cb.exec(cmdFields...) }
		if !isComplete {
			suggestions = cb.stages[stage+1]
			f = nil
		}
		return CommandResult{
			IsCommand:   true,
			IsComplete:  isComplete,
			Suggestions: suggestions,
			Func:        f,
		}
	}

	return CommandResult{
		Error: getFatalCommandErr(cmd),
	}
}

func isCommand(cmds []string, cmd string) bool {
	return lib.StrSliceContains(cmds, cmd)
}

func NewCommandBase(config CommandConfig) CommandBase {
	return CommandBase{
		stages:         config.stages,
		validationFunc: config.validationFunc,
		hasArg:         config.hasArg,
		exec:           config.exec,
	}
}
