package cmd

import (
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib/logger"
)

var (
	ErrFatalCommand    = fmt.Errorf("command parsing failed; this should not happen")
	ErrNotCommand      = fmt.Errorf("unrecognized command")
	ErrIncompleteCmd   = fmt.Errorf("incomplete command")
	ErrMissingArgument = fmt.Errorf("missing argument")
	ErrEmptyCommand    = fmt.Errorf("empty command")
	ErrUnsupportedCmd  = fmt.Errorf("unsupported command chain")
)

var cmdId = 0

type Model interface {
	Update(tea.Msg) (Model, tea.Cmd)
	View() string
	GetErrors() []error
	GetCmdTree() [][]string
	IsSupported(branchStr string) bool
	IsInitialized() bool
	ValidateCommand()
}

type Status struct {
	IsCommand   bool
	IsComplete  bool
	IsSupported bool
	Branches    []string
	Arg         string
	BranchStr   string
	Error       error
}

type Config struct {
	// The overall name of the command
	Name string

	// Command Model which needs to implement the
	// command Base Model.
	Model Model

	// Setting this to true will prevent non-arg
	// commands from working properly
	HasArg bool

	// Validates the command argument. For instance
	// if the user should enter a price, then you
	// would validate that here.
	InputValidationFunc func(arg string) error

	// Limit the character set to be used. For
	// instance, if the argument is supposed to be
	// a price, then don't allow any chars other
	// than numbers and a period.
	KeyValidationFunc func(key rune) bool
}

func New(config Config) Command {
	if config.Name == "" {
		panic("missing command name")
	}

	if config.HasArg && config.InputValidationFunc == nil {
		panic("command arguments need a validation function")
	}

	if config.Model == nil {
		panic("missing command model")
	}

	config.Model.ValidateCommand()

	cmdId += 1
	cmd := Command{
		name:                config.Name,
		model:               config.Model,
		status:              Status{},
		tree:                config.Model.GetCmdTree(),
		inputValidationFunc: config.InputValidationFunc,
		hasArg:              config.HasArg,
		id:                  cmdId,
		keyValidationFunc:   config.KeyValidationFunc,
	}

	return cmd
}

// A command is not a single entry point, but a tree of
// entry points. A single command instance can and will
// allow multiple execution paths via a tree hierarchy.
// The first level of the tree is for command aliases
// and all subsequent levels are leaves attached to
// those aliases.
//
// Example:
//
//	[][]string{{"set"}, {"bill", "stat"}, {"amount", "name"}}
//
// Resulting Command Branches:
//
//	set bill amount
//	set bill name
//	set stat amount
//	set stat name
//
// Each command branch is an execution path. What
// happens in that execution path is up to the dev.
type Command struct {
	id                  int
	name                string
	model               Model
	tree                [][]string
	hasArg              bool
	status              Status
	inputValidationFunc func(arg string) error
	keyValidationFunc   func(key rune) bool
}

func (cb Command) GetId() int {
	return cb.id
}

func (cb *Command) ParseCommand(cmd string) Status {
	cmdFields := strings.Fields(cmd)
	var finalPos int = 0
	var isCommand, isComplete bool

	logger.Log(logger.Insane, "CommandModel", "parsing command [%s]", cmd)

	for pos, cmds := range cb.tree {
		if len(cmdFields) == pos || !slices.Contains(cmds, cmdFields[pos]) {
			break
		}
		finalPos = pos + 1
	}

	isCommand = finalPos > 0

	if isCommand && cb.hasArg && finalPos == len(cb.tree) {
		cs := Status{
			IsCommand:  true,
			IsComplete: true,
			Arg:        cmdFields[len(cmdFields)-1],
			BranchStr:  cmd,
		}
		if len(cmdFields) == finalPos {
			cs.Error = fmt.Errorf("expected a value after '%s'", cmdFields[len(cmdFields)-1])
		} else {
			cs.Error = cb.inputValidationFunc(cmdFields[len(cmdFields)-1])
		}
		return cs
	}

	isComplete = finalPos == len(cmdFields) && !cb.hasArg

	var branches []string
	if finalPos < len(cb.tree) {
		branches = cb.normalizeBranches(cmd, finalPos, cb.tree[finalPos])
	} else {
		branches = cb.normalizeBranches(cmd, finalPos, []string{})
	}

	var err error

	if isCommand && !cb.model.IsSupported(cmd) {
		err = ErrUnsupportedCmd
	}

	if isCommand && !isComplete {
		err = ErrIncompleteCmd
	}

	if !isCommand {
		if cmd == "" {
			err = ErrEmptyCommand
		} else {
			err = ErrNotCommand
		}
	}

	return Status{
		IsCommand:   isCommand,
		IsComplete:  isComplete,
		IsSupported: cb.model.IsSupported(cmd),
		Branches:    branches,
		BranchStr:   cmd,
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
normalizeBranches prepends the previous cmd branch string to the suggestions.
This is necessary because the input box needs the whole phrase as a
completion.
*/
func (cb *Command) normalizeBranches(
	branchStr string,
	treePos int,
	branches []string,
) []string {
	newBranches := make([]string, len(branches))
	copy(newBranches, branches)

	branchStr = strings.TrimSpace(branchStr)
	leaves := strings.Fields(branchStr)

	branchPrefix := ""
	if treePos > 0 && treePos <= len(leaves) {
		branchPrefix = strings.Join(leaves[:treePos], " ") + " "
	}

	// Prevents repeated cmd branches and only allows
	// cmd branches for partially entered cmd branches.
	if len(leaves) == treePos {
		return []string{strings.TrimSpace(branchPrefix)}
	}

	for i, s := range newBranches {
		newBranches[i] = branchPrefix + s
	}

	return newBranches
}
