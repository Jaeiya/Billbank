package cmdmodel

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/jaeiya/billbank/internal/logger"
	"github.com/jaeiya/billbank/internal/ui"
	"github.com/jaeiya/billbank/internal/utils"
)

type CommandData[T any] struct {
	Name     string
	Aliases  []string
	Commands []Command[T]
}

type ArgType int

const (
	ArgNone = ArgType(iota)
	ArgOptional
	ArgRequired
)

type Command[T any] struct {
	// A list of words that execute a specific
	// command function, when entered into the
	// command input. Empty paths refer to
	// the command alias itself as a command.
	Path string

	// The path split by its words without the
	// alias.
	pathParts []string

	// Executes the logic of the command, which
	// updates the command model.
	Run func(T) T

	// The main display function for the command.
	View func(T) string

	// Hides the command input, which relinquishes
	// keyboard control to the command. This is
	// necessary for commands which control the
	// UI using the keyboard.
	CaptureInput bool

	// The type of arguments that the command
	// requires. ArgNone is the default.
	//
	//	ArgNone
	//	ArgOptional
	//	ArgRequired
	ArgType ArgType

	// Parses the argument passed to the command.
	// It should return a user-readable error
	// if it fails to parse.
	//
	// This function is required if the arg type
	// is NOT ArgNone.
	ParseArg func(arg string) (any, error)
}

type Base[T any] struct {
	cmdMap     map[string]Command[T]
	name       string
	aliases    []string
	commands   []Command[T]
	status     Status
	errors     []error
	lastError  error
	id         int
	viewWidth  int
	viewHeight int
}

func NewBaseModel[T any](cmdData CommandData[T]) *Base[T] {
	validateCmdData(cmdData)

	cmdMap := map[string]Command[T]{}
	for i, cmd := range cmdData.Commands {
		leaves := strings.Split(cmd.Path, " ")
		path := strings.Join(leaves, " ")
		for _, alias := range cmdData.Aliases {
			fullCmdPath := strings.TrimSpace(fmt.Sprintf("%s %s", alias, path))
			logger.Log(
				logger.Hot,
				"binding command [%s] to [%s] as [%s]",
				path, alias, fullCmdPath,
			)
			if cmd.Path != "" {
				cmd.pathParts = strings.Split(cmd.Path, " ")
			}
			cmdMap[fullCmdPath] = cmd
			cmdData.Commands[i] = cmd
		}

	}

	logger.Log(
		logger.Info,
		"command [%s] loaded [%d] command paths",
		cmdData.Name,
		len(cmdData.Commands),
	)

	logger.LogFunc(logger.Debug, func() string {
		cmdPaths := make([]string, 0, len(cmdMap))
		for key := range cmdMap {
			cmdPaths = append(cmdPaths, "["+key+"]")
		}
		return "command [" + cmdData.Name + "] loaded " + strings.Join(cmdPaths, ", ")
	})

	return &Base[T]{
		id:        utils.NewID(),
		cmdMap:    cmdMap,
		name:      cmdData.Name,
		aliases:   cmdData.Aliases,
		commands:  cmdData.Commands,
		lastError: fmt.Errorf(""),
	}
}

func (bc *Base[T]) Update(model T, msg tea.Msg) (T, tea.Cmd) {
	var teaCmds []tea.Cmd

	switch msg := msg.(type) {
	case WindowSizeMsg:
		logger.Log(logger.Hot, "window size event triggered [%d:%d]", msg.Width, msg.Height)
		bc.viewHeight = msg.Height
		bc.viewWidth = msg.Width

	case ViewportSizeMsg:
		logger.Log(logger.Hot, "viewport size event triggered [%d:%d]", msg.Width, msg.Height)
		bc.viewHeight = msg.Height
		bc.viewWidth = msg.Width
		if bc.status.Path != "" {
			model = bc.Exec(model)
			logger.Log(logger.Hot, "finished executing [%s]", bc.status.Path)
		}
	}

	return model, tea.Batch(teaCmds...)
}

func (bc *Base[T]) View(model T) string {
	cmdPath := bc.status.Path

	cmd, ok := bc.cmdMap[cmdPath]
	if !ok {
		bc.AddError(fmt.Errorf("could not find command [%s]", bc.status.Path))
	}

	errs := bc.getErrors()
	if len(errs) > 0 {
		return ui.NewErrorBox(
			"Command Error",
			errs[0].Error(),
			bc.viewWidth,
			bc.viewHeight,
		)
	}
	return cmd.View(model)
}

func (m Base[T]) ParseCommand(cmdInput string) Status {
	var inputParts []string = strings.Fields(cmdInput)
	if len(inputParts) == 0 {
		return Status{Error: ErrEmptyCommand}
	}

	var alias string = inputParts[0]
	var cmdPathParts []string = inputParts[1:]
	var cmdPaths []string

	if !slices.Contains(m.aliases, alias) {
		return Status{Error: ErrNotCommand}
	}

	for _, cmd := range m.commands {
		cmdPath := strings.Join(cmdPathParts, " ")
		cmdPaths = append(cmdPaths, fmt.Sprintf("%s %s", alias, cmd.Path))

		if cmd.ArgType < ArgRequired && cmdPath == cmd.Path {
			return Status{
				IsCommand:    true,
				CaptureInput: cmd.CaptureInput,
				Path:         strings.TrimSpace(alias + " " + cmd.Path),
				Arg:          nil,
				Error:        nil,
			}
		}

		hasArg := len(cmdPathParts) == len(cmd.pathParts)+1

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
				v, err := cmd.ParseArg(arg)
				if v == nil && err == nil {
					err = ErrMisconfiguredArgParser
				}
				return Status{
					IsCommand:    true,
					CaptureInput: cmd.CaptureInput,
					Path:         strings.TrimSpace(alias + " " + cmd.Path),
					Arg:          v,
					Error:        err,
				}
			}
		}
	}

	return Status{
		IsCommand:   true,
		Error:       ErrIncompleteCmd,
		Suggestions: populateSuggestions(cmdPaths, inputParts),
		Path:        cmdInput,
	}
}

func (m Base[T]) GetViewSize() (int, int) {
	return m.viewWidth, m.viewHeight
}

func (m Base[T]) GetId() int {
	return m.id
}

func (m Base[T]) GetStatus() Status {
	return m.status
}

func (m *Base[T]) SetStatus(s Status) {
	m.status = s
}

func (m Base[T]) GetName() string {
	return m.name
}

func (m *Base[T]) AddError(err error) {
	m.errors = append(m.errors, err)
}

func (m *Base[T]) AddArgTypeError(arg any, expectedType string) {
	m.errors = append(
		m.errors,
		fmt.Errorf(
			"[%s] has a misconfigured arg type: [%s] expected [%s]",
			m.status.Path,
			reflect.TypeOf(arg),
			expectedType,
		),
	)
}

func (m Base[T]) GetAliases() []string {
	aliases := make([]string, len(m.aliases))
	copy(aliases, m.aliases)
	return aliases
}

func (m Base[T]) GetCmdPaths() (paths []string) {
	paths = make([]string, 0, len(m.cmdMap))
	for k := range m.cmdMap {
		paths = append(paths, k)
	}
	return paths
}

func (m Base[T]) IsActivePath(cmdPath string) bool {
	return m.status.Path == cmdPath
}

// IsInitialized checks to make sure that various expected values
// are set.
func (m Base[T]) IsInitialized() bool {
	b := len(m.cmdMap) > 0 &&
		len(m.status.Path) > 0 &&
		m.viewWidth > 0 &&
		m.viewHeight > 0
	return b
}

// Exec executes the current command path in the context of the
// passed model. All detected errors are logged and stored.
func (m *Base[T]) Exec(model T) T {
	cmdPath := m.status.Path
	cmd := m.cmdMap[cmdPath]

	m.clearErrors()
	logger.Log(logger.Hot, "executing command path [%s]", cmdPath)

	if cmd.Run == nil {
		err := fmt.Errorf("[%s] has an unimplemented Run func()", cmdPath)
		if m.lastError.Error() != err.Error() {
			logger.Log(logger.Error, "%s", err)
			m.lastError = err
			m.AddError(err)
		}
		return model
	}

	if cmd.View == nil {
		err := fmt.Errorf("[%s] has an unimplemented View func()", cmdPath)
		if m.lastError.Error() != err.Error() {
			logger.Log(logger.Error, "%s", err)
			m.lastError = err
			m.AddError(err)
		}
		return model
	}

	model = cmd.Run(model)
	errs := m.getErrors()
	if len(errs) > 0 {
		if m.lastError.Error() != errs[0].Error() {
			logger.Log(logger.Error, "%s", errs[0].Error())
		}
		m.lastError = errs[0]
	}

	return model
}

func (m Base[T]) getErrors() []error {
	return m.errors
}

func (m *Base[T]) clearErrors() {
	if len(m.errors) > 0 {
		m.lastError = fmt.Errorf("")
		m.errors = nil
	}
}

func validateCmdData[T any](cmdData CommandData[T]) {
	if len(cmdData.Commands) == 0 {
		logger.LogFatal(
			"Command [%s] has no command paths",
			"Did you forget to add commands to a new command model?",
			cmdData.Name,
		)
	}

	for _, cmd := range cmdData.Commands {
		leaves := strings.Split(cmd.Path, " ")
		hasAlias := slices.ContainsFunc(cmdData.Aliases, func(alias string) bool {
			return leaves[0] == alias
		})

		if hasAlias {
			logger.LogFatal(
				"[%s] command path [%s] does not need to include the command-alias [%s]",
				MsgUsingAliasInCmdPathErr,
				cmdData.Name, cmd.Path, leaves[0],
			)
		}
	}

	cmdMap := map[string]struct{}{}
	for _, cmd := range cmdData.Commands {
		if cmd.Path == "" && cmd.ArgType > ArgNone && len(cmdData.Commands) > 1 {
			logger.LogFatal(
				"[%s] has been initialized as a default command with args, but also has other command paths",
				MsgIsCmdItselfErr,
				cmdData.Name,
			)
		}

		if cmd.ArgType > ArgNone && cmd.ParseArg == nil {
			logger.LogFatal(
				"[%s] command path [%s] is missing an arg validation function.",
				MsgMissingArgFuncErr,
				cmdData.Name, cmd.Path,
			)
		}
		if _, ok := cmdMap[cmd.Path]; ok {
			logger.LogFatal(
				"[%s] contains a duplicate command path [%s]",
				"Is it possible you were testing something and accidentally duplicated a command?",
				cmdData.Name,
				cmd.Path,
			)
		}
		cmdMap[cmd.Path] = struct{}{}
	}
}

func populateSuggestions(cmdPaths, pathParts []string) []string {
	suggestions := make([]string, 0, 10) // set a reasonable minimum length
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
