package cmd

import (
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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

	// Validates the command argument. For instance
	// if the user should enter a price, then you
	// would validate that here.
}

func New(config Config) Command {
	if config.Name == "" {
		panic("missing command name")
	}

	if config.Model == nil {
		panic("missing command model")
	}

	config.Model.ValidateCommand()

	cmdId += 1
	cmd := Command{
		name:   config.Name,
		model:  config.Model,
		status: Status{},
		tree:   config.Model.GetCmdTree(),
		id:     cmdId,
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
	id     int
	name   string
	model  Model
	tree   Tree
	status Status
}

type Tree struct {
	Aliases  []string
	Branches []Branch
}

type Branch struct {
	Leaves []string
	HasArg bool
	// Allows you to validate the users input
	// argument before the command is run.
	ValidateArg func(arg string) error
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

		if !b.HasArg && inputPath == cmdPath && cb.model.IsSupported(cmdPathInput) {
			return Status{
				IsCommand: true,
				Path:      strings.TrimSpace(alias + " " + cmdPath),
			}
		}

		// Is the command valid, but missing an argument?
		if b.HasArg && len(b.Leaves) == len(pathParts)-1 && inputPath == cmdPath {
			return Status{
				IsCommand: true,
				Error: fmt.Errorf(
					"expected value after '%s'",
					pathParts[len(pathParts)-1],
				),
			}
		}

		if b.HasArg && len(pathParts) == len(b.Leaves)+2 {
			// Do we have a valid path when excluding the argument?
			inputPath = strings.Join(pathParts[1:len(pathParts)-1], " ")
			if inputPath == cmdPath {
				err = activeBranch.ValidateArg(pathParts[len(pathParts)-1])
				return Status{
					IsCommand: true,
					Path:      strings.TrimSpace(alias + " " + inputPath),
					Arg:       pathParts[len(pathParts)-1],
					Error:     err,
				}
			}
		}
	}

	return Status{
		IsCommand:       true,
		Error:           ErrIncompleteCmd,
		PathSuggestions: populateSuggestions(cmdPaths, pathParts),
		Path: strings.TrimSpace(
			fmt.Sprintf("%s %s", pathParts[0], strings.Join(activeBranch.Leaves, " ")),
		),
	}
}

func populateSuggestions(cmdPaths []string, pathParts []string) []string {
	var suggestions []string
	for _, path := range cmdPaths {
		leaves := strings.Fields(path)
		if len(pathParts) == len(leaves) {
			suggestions = append(
				suggestions,
				strings.Join(leaves[:len(pathParts)], " "),
			)
		}
	}
	return suggestions
}
