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

func NewCommandBase(config CommandConfig) CommandBase {
	return CommandBase{
		stages:         config.stages,
		validationFunc: config.validationFunc,
		hasArg:         config.hasArg,
		exec:           config.exec,
	}
}

func (cb *CommandBase) GetStage(stage int) []string {
	if stage >= len(cb.stages) || stage < 0 {
		panic("stage does not exist")
	}
	return cb.stages[stage]
}

func (cb *CommandBase) ParseCommand(cmd string) CommandResult {
	cmdFields := strings.Fields(cmd)

	if len(cmdFields) > len(cb.stages) && !cb.hasArg {
		return CommandResult{Error: ErrCommandNotFound}
	}

	for stage, cmds := range cb.stages {
		if stage == len(cmdFields) || !isCommand(cmds, cmdFields[stage]) {
			var isCommand bool
			suggestions := cb.stages[0]
			if stage > 0 {
				isCommand = true
				suggestions = cb.stages[stage]
			}
			return CommandResult{
				IsCommand:   isCommand,
				Suggestions: cb.normalizeSuggestions(cmd, suggestions),
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
		suggestions := cb.stages[0]
		f := func() { cb.exec(cmdFields...) }
		if !isComplete {
			suggestions = cb.stages[stage+1]
			f = nil
		}
		return CommandResult{
			IsCommand:   true,
			IsComplete:  isComplete,
			Suggestions: cb.normalizeSuggestions(cmd, suggestions),
			Func:        f,
		}
	}

	return CommandResult{
		Error: getFatalCommandErr(cmd),
	}
}

/*
normalizeSuggestions prepends the previous command string to the suggestions.
This is necessary because the input box needs the whole phrase as a
completion.
*/
func (cb *CommandBase) normalizeSuggestions(cmd string, suggestions []string) []string {
	normSuggestions := make([]string, len(suggestions))
	copy(normSuggestions, suggestions)

	if spcIndex := strings.LastIndex(cmd, " "); spcIndex != -1 {
		prefix := strings.TrimSpace(cmd[:spcIndex])
		for i, s := range normSuggestions {
			normSuggestions[i] = prefix + " " + s
		}
	}

	return normSuggestions
}

func isCommand(cmds []string, cmd string) bool {
	return lib.StrSliceContains(cmds, cmd)
}
