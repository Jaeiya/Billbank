package cmd

import (
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib/logger"
)

var (
	ErrFatalCommand     = fmt.Errorf("command parsing failed; this should not happen")
	ErrNotCommand       = fmt.Errorf("unrecognized command")
	ErrIncompleteCmd    = fmt.Errorf("incomplete command")
	ErrMissingArgument  = fmt.Errorf("missing argument")
	ErrEmptyCommand     = fmt.Errorf("empty command")
	ErrUnimplementedCmd = fmt.Errorf("unimplemented command")
)

var cmdId = 0

type Model interface {
	Update(tea.Msg) (Model, tea.Cmd)
	View() string
	GetErrors() []error
	GetCmdTree() Tree
	IsSupported(branchStr string) bool
	IsInitialized() bool
	ValidateCommand()
}

type Status struct {
	IsCommand         bool
	BranchSuggestions []string
	Arg               string
	BranchStr         string
	Error             error
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
	tree                Tree
	hasArg              bool
	status              Status
	inputValidationFunc func(arg string) error
	keyValidationFunc   func(key rune) bool
}

type Tree struct {
	Aliases  []string
	Branches []CmdBranch
}

type CmdBranch struct {
	Leaves []string
	HasArg bool
}

func (cb Command) GetId() int {
	return cb.id
}

func (cb Command) ParseCommand(input string) Status {
	var cmdFields []string = strings.Fields(input)
	if len(cmdFields) == 0 {
		return Status{Error: ErrEmptyCommand}
	}

	var alias string = cmdFields[0]
	var isCompleted bool
	var err error
	var branches []string
	var activeBranch CmdBranch

	if !slices.Contains(cb.tree.Aliases, alias) {
		return Status{Error: ErrNotCommand}
	}

	for _, b := range cb.tree.Branches {
		fieldStr := strings.Join(cmdFields[1:], " ")
		branchStr := strings.Join(b.Leaves, " ")
		branches = append(branches, fmt.Sprintf("%s %s", alias, branchStr))
		activeBranch = b

		if b.HasArg {
			if len(b.Leaves) == len(cmdFields)-1 {
				if fieldStr == branchStr {
					isCompleted = false
					break
				}
			}

			if len(cmdFields) == len(b.Leaves)+2 {
				fieldStr = strings.Join(cmdFields[1:len(cmdFields)-1], " ")
				if fieldStr == branchStr {
					isCompleted = true
					break
				}
			}
		}

		if fieldStr == branchStr {
			isCompleted = true
			break
		}

	}

	if !isCompleted {
		err = ErrIncompleteCmd
	}

	if !isCompleted && activeBranch.HasArg && len(activeBranch.Leaves) == len(cmdFields)-1 {
		err = fmt.Errorf("expected value after %s ", cmdFields[len(cmdFields)-1])
	}

	if isCompleted && !cb.model.IsSupported(input) {
		err = ErrUnimplementedCmd
	}

	var arg string
	if activeBranch.HasArg && isCompleted {
		arg = cmdFields[len(cmdFields)-1]
	}

	var suggestedBranches []string
	for _, b := range branches {
		leaves := strings.Fields(b)
		if len(cmdFields) == len(leaves) {
			suggestedBranches = append(
				suggestedBranches,
				strings.Join(leaves[:len(cmdFields)], " "),
			)
		}
	}

	logger.Log(logger.Insane, "CommandModel", "branch suggestions [%+v]", suggestedBranches)

	branchStr := fmt.Sprintf("%s %s", cmdFields[0], strings.Join(activeBranch.Leaves, " "))
	if len(activeBranch.Leaves) == 0 {
		branchStr = cmdFields[0]
	}

	return Status{
		IsCommand:         true,
		BranchStr:         branchStr,
		BranchSuggestions: suggestedBranches,
		Arg:               arg,
		Error:             err,
	}
}

func (cb *Command) ValidateKey(key rune) bool {
	if cb.keyValidationFunc != nil && cb.hasArg {
		return cb.keyValidationFunc(key)
	}
	return true
}
