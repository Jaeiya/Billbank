package commands

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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

type CommandStatus struct {
	IsCommand   bool
	IsComplete  bool
	Suggestions []string
	TreePos int
	Error   error
}

type CommandConfig struct {
	Command
}

type Command struct {
	// Represents the way a command is hierarchically constructed
	// including aliases.
	//
	// Commands Example:
	//		set bill amount
	//		set bill name
	//		set stat amount
	// 		set stat name
	// Commands Structure:
	//		[][]string{{"set"}, {"bill", "stat"}, {"amount", "name"}}
	tree                [][]string
	hasArg              bool
	execFunc            func(args ...string) tea.Model
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

func (cb *Command) ParseCommand(cmd string) CommandStatus {
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
		return CommandStatus{
			IsCommand:  true,
			IsComplete: true,
			Error:      cb.inputValidationFunc(cmdFields[len(cmdFields)-1]),
			TreePos:    finalPos,
		}
	}

	isComplete = len(cmdFields) == finalPos && !cb.hasArg

	var suggestions []string
	if finalPos < len(cb.tree) {
		suggestions = cb.normalizeSuggestions(cmd, finalPos, cb.tree[finalPos])
	} else {
		suggestions = cb.normalizeSuggestions(cmd, finalPos, []string{})
	}

	var err error
	if isCommand && !isComplete {
		err = ErrInvalidCommand
	}

	if !isCommand {
		err = ErrNotCommand
	}

	return CommandStatus{
		IsCommand:   isCommand,
		IsComplete:  isComplete,
		Suggestions: suggestions,
		TreePos:     finalPos,
		Error:       err,
	}
}

func (cb *Command) ValidateKey(key rune) bool {
	if cb.keyValidationFunc != nil && cb.hasArg {
		return cb.keyValidationFunc(key)
	}
	return true
}

func (cb Command) Execute() {
	cb.execFunc()
}

/*
normalizeSuggestions prepends the previous command string to the suggestions.
This is necessary because the input box needs the whole phrase as a
completion.
*/
func (cb *Command) normalizeSuggestions(
	cmd string,
	treePos int,
	suggestions []string,
) []string {
	normSuggestions := make([]string, len(suggestions))
	copy(normSuggestions, suggestions)

	cmd = strings.TrimSpace(cmd)
	cmdParts := strings.Fields(cmd)

	cmdPrefix := ""
	if treePos > 0 && treePos <= len(cmdParts) {
		cmdPrefix = strings.Join(cmdParts[:treePos], " ") + " "
	}

	// Prevents repeated suggestions and only allows
	// suggestions for partially entered commands.
	if len(cmdParts) == treePos {
		return []string{strings.TrimSpace(cmdPrefix)}
	}

	for i, s := range normSuggestions {
		normSuggestions[i] = cmdPrefix + s
	}

	return normSuggestions
}
