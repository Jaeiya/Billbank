package cmdmodel

import (
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	ErrNotCommand       = fmt.Errorf("unrecognized command")
	ErrIncompleteCmd    = fmt.Errorf("incomplete command")
	ErrMissingArgument  = fmt.Errorf("missing argument")
	ErrEmptyCommand     = fmt.Errorf("empty command")
	ErrUnimplementedCmd = fmt.Errorf("unimplemented command")
)

var cmdId = 0

type Interface interface {
	Update(tea.Msg) (Interface, tea.Cmd)
	View() string
	GetName() string
	GetErrors() []error
	GetCmdData() CommandData
	IsSupported(cmdPath string) bool
	IsInitialized() bool
}

type Status struct {
	IsCommand       bool
	PathSuggestions []string
	Path            string
	Arg             string
	Error           error
}

type CommandData struct {
	Aliases  []string
	Commands []ModelCommand
}

func newModelCommand[T any](baseCmd BaseCommand[T]) ModelCommand {
	return ModelCommand{
		baseCmd.Path,
		strings.Split(baseCmd.Path, " "),
		baseCmd.ArgType,
		baseCmd.ValidateArg,
	}
}

type ModelCommand struct {
	Path        string
	PathParts   []string
	ArgType     ArgType
	ValidateArg func(arg string) error
}

func New(model Interface) Model {
	if model == nil {
		panic("missing command model")
	}
	cmdId += 1
	return Model{cmdId, model, Status{}}
}

type Model struct {
	id     int
	model  Interface
	status Status
}

func (cb Model) GetId() int {
	return cb.id
}

func (cb Model) ParseCommand(cmdInput string) Status {
	var inputParts []string = strings.Fields(cmdInput)
	if len(inputParts) == 0 {
		return Status{Error: ErrEmptyCommand}
	}

	var alias string = inputParts[0]
	var cmdPathParts []string = inputParts[1:]
	var cmdPaths []string

	cmdData := cb.model.GetCmdData()

	if !slices.Contains(cmdData.Aliases, alias) {
		return Status{Error: ErrNotCommand}
	}

	for _, cmd := range cmdData.Commands {
		cmdPath := strings.Join(cmdPathParts, " ")
		cmdPaths = append(cmdPaths, fmt.Sprintf("%s %s", alias, cmd.Path))

		if cmd.ArgType < ArgRequired && cmdPath == cmd.Path {
			var err error
			if !cb.model.IsSupported(cmdInput) {
				err = ErrUnimplementedCmd
			}
			return Status{
				IsCommand: true,
				Path:      strings.TrimSpace(alias + " " + cmd.Path),
				Error:     err,
			}
		}

		hasArg := len(cmdPathParts) == len(cmd.PathParts)+1

		if cmd.ArgType > ArgNone && !hasArg && cmdPath == cmd.Path {
			return Status{
				IsCommand: true,
				Error: fmt.Errorf(
					"expected value after '%s'",
					inputParts[len(inputParts)-1],
				),
			}
		}

		if cmd.ArgType > ArgNone && hasArg {
			cmdPath = strings.Join(cmdPathParts[:len(cmdPathParts)-1], " ")
			if cmdPath == cmd.Path {
				arg := inputParts[len(inputParts)-1]
				err := cmd.ValidateArg(arg)
				return Status{
					IsCommand: true,
					Path:      strings.TrimSpace(alias + " " + cmd.Path),
					Arg:       inputParts[len(inputParts)-1],
					Error:     err,
				}
			}
		}
	}

	return Status{
		IsCommand:       true,
		Error:           ErrIncompleteCmd,
		PathSuggestions: populateSuggestions(cmdPaths, inputParts),
		Path:            cmdInput,
	}
}

func populateSuggestions(cmdPaths, pathParts []string) []string {
	var suggestions []string
	for _, path := range cmdPaths {
		cmdPathParts := strings.Fields(path)
		if len(cmdPathParts) >= len(pathParts) {
			suggestions = append(
				suggestions,
				strings.Join(cmdPathParts[:len(pathParts)], " "),
			)
		}
	}
	return suggestions
}
