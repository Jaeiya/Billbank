package ui

import (
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type CommandError error

var (
	ErrFatalCommand = CommandError(
		fmt.Errorf("command parsing failed; this should not happen"),
	)
	ErrNotCommand      = CommandError(fmt.Errorf("unrecognized command"))
	ErrInvalidCommand  = CommandError(fmt.Errorf("incomplete command"))
	ErrMissingArgument = CommandError(fmt.Errorf("missing argument"))
)

var cmdId = 0

type CommandStatus struct {
	IsCommand   bool
	IsComplete  bool
	Suggestions []string
	Arg         string
	CommandStr  string
	TreeStr     string
	// The command tree position of the input. A command can be in an incomplete state,
	// which means the input is correct, but it's in a lower position within the
	// command tree hierarchy.
	//
	// Example:
	//		set              // TreePos 0
	// 		set bills        // TreePos 1
	//		set bills amount // TreePos 2
	TreePos int
	Error   error
}

type CommandConfig struct {
	Model               CommandModel
	Tree                [][]string
	HasArg              bool
	InputValidationFunc func(arg string) error
	KeyValidationFunc   func(key rune) bool
}

type CommandModel interface {
	Update(tea.Msg) (CommandModel, tea.Cmd)
	View() string
	GetLastError() error
}

type CommandMsg Command

type Command struct {
	CommandModel
	// Represents the way a command is hierarchically constructed
	// including aliases.
	//
	// Example:
	//		[][]string{{"set"}, {"bill", "stat"}, {"amount", "name"}}
	// Resulting Commands:
	//		set bill amount
	//		set bill name
	//		set stat amount
	// 		set stat name
	id                  int
	tree                [][]string
	hasArg              bool
	status              CommandStatus
	inputValidationFunc func(arg string) error
	keyValidationFunc   func(key rune) bool
}

func NewCommand(config CommandConfig) Command {
	if config.HasArg && config.InputValidationFunc == nil {
		panic("command arguments need a validation function")
	}
	cmdId += 1
	cmd := Command{
		CommandModel:        config.Model,
		status:              CommandStatus{},
		tree:                config.Tree,
		inputValidationFunc: config.InputValidationFunc,
		hasArg:              config.HasArg,
		id:                  cmdId,
		keyValidationFunc:   config.KeyValidationFunc,
	}

	return cmd
}

func (cb Command) GetAliases() []string {
	return cb.tree[0]
}

func (cb Command) GetId() int {
	return cb.id
}

func (cb *Command) ParseCommand(cmd string) CommandStatus {
	cmdFields := strings.Fields(cmd)
	var finalPos int = 0
	var isCommand, isComplete bool
	var cmdStr string

	for pos, cmds := range cb.tree {
		if len(cmdFields) == pos || !slices.Contains(cmds, cmdFields[pos]) {
			break
		}
		finalPos = pos + 1
		cmdStr = cmdFields[pos]
	}

	isCommand = finalPos > 0

	if isCommand && cb.hasArg && finalPos == len(cb.tree) {
		cs := CommandStatus{
			IsCommand:  true,
			IsComplete: true,
			Arg:        cmdFields[len(cmdFields)-1],
			TreePos:    finalPos - 1,
			CommandStr: cmdStr,
			TreeStr:    cmd,
		}
		if len(cmdFields) == finalPos {
			cs.Error = fmt.Errorf("expected a value after '%s'", cmdFields[len(cmdFields)-1])
		} else {
			cs.Error = cb.inputValidationFunc(cmdFields[len(cmdFields)-1])
		}
		return cs
	}

	isComplete = finalPos == len(cb.tree) && !cb.hasArg

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
		TreePos:     finalPos - 1,
		CommandStr:  cmdStr,
		TreeStr:     cmd,
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
