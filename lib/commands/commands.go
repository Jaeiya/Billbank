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
	ErrNotCommand      = CommandError(fmt.Errorf("unrecognized command"))
	ErrInvalidCommand  = CommandError(fmt.Errorf("command is formatted incorrectly"))
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
	Stage       int
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
	if config.hasArg && config.validationFunc == nil {
		panic("command arguments need a validation function")
	}
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
	var finalStage int = 0
	var isCommand, isComplete bool

	for stage, cmds := range cb.stages {
		if len(cmdFields) == stage || !lib.StrSliceContains(cmds, cmdFields[stage]) {
			break
		}
		finalStage = stage + 1
	}

	isCommand = finalStage > 0

	if isCommand && cb.hasArg && finalStage == len(cb.stages) {
		return CommandResult{
			IsCommand:  true,
			IsComplete: true,
			Func:       func() { cb.exec(cmdFields...) },
			Error:      cb.validationFunc(cmdFields[len(cmdFields)-1]),
			Stage:      finalStage,
		}
	}

	isComplete = len(cmdFields) == finalStage && !cb.hasArg

	suggestions := cb.normalizeSuggestions(cmd, finalStage, []string{})
	if finalStage < len(cb.stages) {
		suggestions = cb.normalizeSuggestions(cmd, finalStage, cb.stages[finalStage])
	}

	var err error
	if isCommand && !isComplete {
		err = ErrInvalidCommand
	}

	if !isCommand {
		err = ErrNotCommand
	}

	return CommandResult{
		IsCommand:   isCommand,
		IsComplete:  isComplete,
		Suggestions: suggestions,
		Stage:       finalStage,
		Error:       err,
	}
}

/*
normalizeSuggestions prepends the previous command string to the suggestions.
This is necessary because the input box needs the whole phrase as a
completion.
*/
func (cb *CommandBase) normalizeSuggestions(
	cmd string,
	stage int,
	suggestions []string,
) []string {
	normSuggestions := make([]string, len(suggestions))
	copy(normSuggestions, suggestions)

	cmd = strings.TrimSpace(cmd)
	cmdParts := strings.Split(cmd, " ")

	cmdPrefix := ""
	if stage > 0 && stage <= len(cmdParts) {
		cmdPrefix = strings.Join(cmdParts[:stage], " ") + " "
	}

	// Prevents repeated suggestions and only allows
	// suggestions for partially entered commands.
	if len(cmdParts) == stage {
		return []string{strings.TrimSpace(cmdPrefix)}
	}

	for i, s := range normSuggestions {
		normSuggestions[i] = cmdPrefix + s
	}

	return normSuggestions
}
