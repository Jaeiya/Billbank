package cmdmodel

import (
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib/logger"
)

var (
	ErrNotCommand       = fmt.Errorf("unrecognized command")
	ErrIncompleteCmd    = fmt.Errorf("incomplete command")
	ErrMissingArgument  = fmt.Errorf("missing argument")
	ErrEmptyCommand     = fmt.Errorf("empty command")
	ErrUnimplementedCmd = fmt.Errorf("unimplemented command")
)

var MsgDuplicateBranchErr = `
Did you forget to remove some test branches? Commands can only
contain tree branches with unique leaf combinations. For instance,
the leaves "hello" & "world" have two unique combinations. You
can have two branches, one with "hello world" and one with
"world hello", but not more than one of each.
`

var MsgNoBranchesWithDefaultArgErr = `
Your command alias directly requires an argument, which means all
branches other than the default branch, are hidden. Consider
turning the command into a compound command: <alias keyword arg>
instead of: <alias arg>`

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
	Commands []Command
}

type Command struct {
	Leaves  []string
	NeedArg bool
	// Allows you to validate the users input
	// argument before the command is run.
	ValidateArg func(arg string) error
}

func New(model Interface) Model {
	if model == nil {
		panic("missing command model")
	}

	tree := model.GetCmdData()

	branchNameStore := map[string]struct{}{}
	for _, branch := range tree.Commands {
		if len(branch.Leaves) == 0 && len(tree.Commands) > 1 && branch.NeedArg {
			logger.LogFatal(
				"Branches on the [%s] command have been hidden implicitly.",
				MsgNoBranchesWithDefaultArgErr,
				tree.Aliases[0],
			)
		}
		cmdPath := strings.Join(branch.Leaves, " ")
		if _, ok := branchNameStore[cmdPath]; ok {
			logger.LogFatal(
				"Found duplicate tree branches [%s]. ",
				MsgDuplicateBranchErr,
				tree.Aliases[0]+" "+cmdPath,
			)
		}
		branchNameStore[cmdPath] = struct{}{}
	}

	cmdId += 1
	cmd := Model{cmdId, model, Status{}}

	return cmd
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
	var pathParts []string = strings.Fields(cmdInput)
	if len(pathParts) == 0 {
		return Status{Error: ErrEmptyCommand}
	}

	var alias string = pathParts[0]
	var cmdLeaves []string = pathParts[1:]
	var cmdPaths []string
	var activeBranch Command

	entry := cb.model.GetCmdData()

	if !slices.Contains(entry.Aliases, alias) {
		return Status{Error: ErrNotCommand}
	}

	for _, branch := range entry.Commands {
		inputPath := strings.Join(pathParts[1:], " ")
		cmdPath := strings.Join(branch.Leaves, " ")
		cmdPaths = append(cmdPaths, fmt.Sprintf("%s %s", alias, cmdPath))
		activeBranch = branch

		if !branch.NeedArg && inputPath == cmdPath {
			var err error
			if !cb.model.IsSupported(cmdInput) {
				err = ErrUnimplementedCmd
			}
			return Status{
				IsCommand: true,
				Path:      strings.TrimSpace(alias + " " + cmdPath),
				Error:     err,
			}
		}

		hasArg := len(cmdLeaves) == len(branch.Leaves)+1

		if branch.NeedArg && !hasArg && inputPath == cmdPath {
			return Status{
				IsCommand: true,
				Error: fmt.Errorf(
					"expected value after '%s'",
					pathParts[len(pathParts)-1],
				),
			}
		}

		if branch.NeedArg && hasArg {
			inputPath = strings.Join(cmdLeaves[:len(cmdLeaves)-1], " ")
			if inputPath == cmdPath {
				return Status{
					IsCommand: true,
					Path:      strings.TrimSpace(alias + " " + inputPath),
					Arg:       pathParts[len(pathParts)-1],
					Error:     activeBranch.ValidateArg(pathParts[len(pathParts)-1]),
				}
			}
		}
	}

	return Status{
		IsCommand:       true,
		Error:           ErrIncompleteCmd,
		PathSuggestions: populateSuggestions(cmdPaths, pathParts),
		Path:            cmdInput,
	}
}

func populateSuggestions(cmdPaths, pathParts []string) []string {
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
