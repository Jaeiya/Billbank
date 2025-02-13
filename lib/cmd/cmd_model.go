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
	IsSupported(cmdPath string) bool
	IsInitialized() bool
	ValidateCommand()
}

type Status struct {
	IsCommand       bool
	PathSuggestions []string
	Path            string
	Arg             string
	Error           error
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
		name:          config.Name,
		model:         config.Model,
		status:        Status{},
		tree:          config.Model.GetCmdTree(),
		validateInput: config.InputValidationFunc,
		hasArg:        config.HasArg,
		id:            cmdId,
		validateKey:   config.KeyValidationFunc,
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
	id            int
	name          string
	model         Model
	tree          Tree
	hasArg        bool
	status        Status
	validateInput func(arg string) error
	validateKey   func(key rune) bool
}

type Tree struct {
	Aliases  []string
	Branches []Branch
}

type Branch struct {
	Leaves        []string
	HasArg        bool
	ValidateKey   func(key rune) bool
	ValidateInput func(arg string) error
}

func (cb Command) GetId() int {
	return cb.id
}

func (cb Command) ParseCommand(cmdPathInput string) Status {
	var pathParts []string = strings.Fields(cmdPathInput)
	if len(pathParts) == 0 {
		return Status{Error: ErrEmptyCommand}
	}

	var alias string = pathParts[0]
	var isCompleted bool
	var err error
	var cmdPaths []string
	var activeBranch Branch

	if !slices.Contains(cb.tree.Aliases, alias) {
		return Status{Error: ErrNotCommand}
	}

	for _, b := range cb.tree.Branches {
		inputPath := strings.Join(pathParts[1:], " ")
		cmdPath := strings.Join(b.Leaves, " ")
		cmdPaths = append(cmdPaths, fmt.Sprintf("%s %s", alias, cmdPath))
		activeBranch = b

		if b.HasArg {
			if len(b.Leaves) == len(pathParts)-1 {
				if inputPath == cmdPath {
					isCompleted = false
					break
				}
			}

			if len(pathParts) == len(b.Leaves)+2 {
				inputPath = strings.Join(pathParts[1:len(pathParts)-1], " ")
				if inputPath == cmdPath {
					isCompleted = true
					if cb.validateInput != nil {
						err = activeBranch.ValidateInput(pathParts[len(pathParts)-1])
						return Status{
							IsCommand: true,
							Path:      strings.TrimSpace(alias + " " + inputPath),
							Arg:       pathParts[len(pathParts)-1],
							Error:     err,
						}
					}
					break
				}
			}
		}

		if inputPath == cmdPath {
			isCompleted = true
			break
		}

	}

	if !isCompleted {
		err = ErrIncompleteCmd
	}

	if !isCompleted && activeBranch.HasArg && len(activeBranch.Leaves) == len(pathParts)-1 {
		err = fmt.Errorf("expected value after %s ", pathParts[len(pathParts)-1])
	}

	if isCompleted && !cb.model.IsSupported(cmdPathInput) {
		err = ErrUnimplementedCmd
	}

	var arg string
	if activeBranch.HasArg && isCompleted {
		arg = pathParts[len(pathParts)-1]
	}

	var possiblePaths []string
	// Suggestions are only relevant when we don't know
	// what the command is.
	if !isCompleted {
		for _, b := range cmdPaths {
			leaves := strings.Fields(b)
			if len(pathParts) == len(leaves) {
				possiblePaths = append(
					possiblePaths,
					strings.Join(leaves[:len(pathParts)], " "),
				)
			}
		}
	}

	logger.Log(logger.Insane, "CommandModel", "path suggestions [%+v]", possiblePaths)

	path := fmt.Sprintf("%s %s", pathParts[0], strings.Join(activeBranch.Leaves, " "))
	if len(activeBranch.Leaves) == 0 {
		path = pathParts[0]
	}

	return Status{
		IsCommand:       true,
		Error:           err,
		PathSuggestions: possiblePaths,
		Path:            path,
		Arg:             arg,
	}
}

func (cb *Command) ValidateKey(key rune) bool {
	if cb.validateKey != nil && cb.hasArg {
		return cb.validateKey(key)
	}
	return true
}
