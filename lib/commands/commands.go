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

type CommandResult struct {
	IsCommand   bool
	IsComplete  bool
	Suggestions []string
	Pos         int
	Error       error
	Func        func()
}

type CommandConfig struct {
	Command
}

type Command struct {
	// Command structure including alias hierarchy. Example:
	//		set bill amount
	//		set bill name
	//		set stat amount
	// 		set stat name
	// Command Structure:
	//		[][]string{{"set"}, {"bill", "stat"}, {"amount", "name"}}
	tree                [][]string
	hasArg              bool
	execFunc            func(args ...string)
	inputValidationFunc func(arg string) error
	keyValidationFunc   func(key rune) bool
}

func NewCommand(config CommandConfig) Command {
	if config.hasArg && config.inputValidationFunc == nil {
		panic("command arguments need a validation function")
	}
	return Command{
		tree:                config.tree,
		inputValidationFunc: config.inputValidationFunc,
		hasArg:              config.hasArg,
		execFunc:            config.execFunc,
		keyValidationFunc:   config.keyValidationFunc,
	}
}

func (cb *Command) GetPosition(pos int) []string {
	if pos >= len(cb.tree) || pos < 0 {
		panic("stage does not exist")
	}
	return cb.tree[pos]
}

func (cb *Command) ParseCommand(cmd string) CommandResult {
	cmdFields := strings.Fields(cmd)
	var finalPos int = 0
	var isCommand, isComplete bool

	for pos, cmds := range cb.tree {
		if len(cmdFields) == pos || !lib.StrSliceContains(cmds, cmdFields[pos]) {
			break
		}
		finalPos = pos + 1
	}

	isCommand = finalPos > 0

	if isCommand && cb.hasArg && finalPos == len(cb.tree) {
		return CommandResult{
			IsCommand:  true,
			IsComplete: true,
			Func:       func() { cb.execFunc(cmdFields...) },
			Error:      cb.inputValidationFunc(cmdFields[len(cmdFields)-1]),
			Pos:        finalPos,
		}
	}

	isComplete = len(cmdFields) == finalPos && !cb.hasArg

	suggestions := cb.normalizeSuggestions(cmd, finalPos, []string{})
	if finalPos < len(cb.tree) {
		suggestions = cb.normalizeSuggestions(cmd, finalPos, cb.tree[finalPos])
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
		Pos:         finalPos,
		Error:       err,
	}
}

func (cb *Command) ValidateKey(key rune) bool {
	if cb.keyValidationFunc != nil && cb.hasArg {
		return cb.keyValidationFunc(key)
	}
	return true
}

/*
normalizeSuggestions prepends the previous command string to the suggestions.
This is necessary because the input box needs the whole phrase as a
completion.
*/
func (cb *Command) normalizeSuggestions(
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
