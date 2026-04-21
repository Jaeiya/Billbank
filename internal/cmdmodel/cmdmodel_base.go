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

// Command is the base struct for all commands and expects a
// tea model as T.
type Command[T any] struct {
	// A word or string of words separated by a space, which lead to
	// the execution of the command.
	//
	// Ex: "clear" or "clear log" or "clear history"
	//
	// 🟡 An empty path refers to the command alias itself as the path.
	Path string

	// A slice containing each word found in the path
	//
	// 🟡 Will be empty if the path is left empty
	pathParts []string

	// Acts as the commands Update func inside of a tea model
	Run func(T) T

	// Acts as the commands View func inside of a tea model
	View func(T) tea.View

	// Hides the command input, which relinquishes keyboard control to
	// the command. This is necessary for commands which control the
	// UI using the keyboard.
	CaptureInput bool

	// The type of arguments that the command requires.
	// ArgNone is the default.
	ArgType ArgType

	// Parses the argument passed to the command. It should return a
	// user-readable error if it fails to parse.
	//
	// 🟠 This function is required if the arg type is NOT ArgNone.
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

func (b *Base[T]) Update(model T, msg tea.Msg) (T, tea.Cmd) {
	var teaCmds []tea.Cmd

	switch msg := msg.(type) {
	case WindowSizeMsg:
		logger.Log(logger.Hot, "window size event triggered [%d:%d]", msg.Width, msg.Height)
		b.viewHeight = msg.Height
		b.viewWidth = msg.Width

	// Sent every time a command is updated and therefore
	// is responsible for command execution.
	case ViewportSizeMsg:
		logger.Log(logger.Hot, "viewport size event triggered [%d:%d]", msg.Width, msg.Height)
		b.viewHeight = msg.Height
		b.viewWidth = msg.Width
		// Make sure a command is active before execution
		if b.status.Path != "" {
			model = b.Exec(model)
			logger.Log(logger.Hot, "finished executing [%s]", b.status.Path)
		}
	}

	return model, tea.Batch(teaCmds...)
}

func (b *Base[T]) View(model T) tea.View {
	cmdPath := b.status.Path
	v := tea.NewView("")

	cmd, ok := b.cmdMap[cmdPath]
	if !ok {
		b.AddError(fmt.Errorf("could not find command [%s]", b.status.Path))
	}

	errs := b.getErrors()
	if len(errs) > 0 {
		v.SetContent(ui.NewErrorBox(
			"Command Error",
			errs[0].Error(),
			b.viewWidth,
			b.viewHeight,
		))
		return v
	}

	return cmd.View(model)
}

func (b Base[T]) ParseCommand(cmdInput string) Status {
	var inputParts []string = strings.Fields(cmdInput)
	if len(inputParts) == 0 {
		return Status{Error: ErrEmptyCommand}
	}

	var alias string = inputParts[0]
	var cmdPathParts []string = inputParts[1:]
	var cmdPaths []string

	if !slices.Contains(b.aliases, alias) {
		return Status{Error: ErrNotCommand}
	}

	for _, cmd := range b.commands {
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

func (b Base[T]) GetViewSize() (int, int) {
	return b.viewWidth, b.viewHeight
}

func (b Base[T]) GetId() int {
	return b.id
}

func (b Base[T]) GetStatus() Status {
	return b.status
}

func (b *Base[T]) SetStatus(s Status) {
	b.status = s
}

func (b Base[T]) GetName() string {
	return b.name
}

func (b *Base[T]) AddError(err error) {
	b.errors = append(b.errors, err)
}

func (b *Base[T]) AddArgTypeError(arg any, expectedType string) {
	b.errors = append(
		b.errors,
		fmt.Errorf(
			"[%s] has a misconfigured arg type: [%s] expected [%s]",
			b.status.Path,
			reflect.TypeOf(arg),
			expectedType,
		),
	)
}

func (b Base[T]) GetAliases() []string {
	aliases := make([]string, len(b.aliases))
	copy(aliases, b.aliases)
	return aliases
}

func (b Base[T]) GetCmdPaths() (paths []string) {
	paths = make([]string, 0, len(b.cmdMap))
	for k := range b.cmdMap {
		paths = append(paths, k)
	}
	return paths
}

func (b Base[T]) IsActivePath(cmdPath string) bool {
	return b.status.Path == cmdPath
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
func (b *Base[T]) Exec(model T) T {
	cmdPath := b.status.Path
	cmd := b.cmdMap[cmdPath]

	b.clearErrors()
	logger.Log(logger.Hot, "executing command path [%s]", cmdPath)

	if cmd.Run == nil {
		err := fmt.Errorf("[%s] has an unimplemented Run func()", cmdPath)
		if b.lastError.Error() != err.Error() {
			logger.Log(logger.Error, "%s", err)
			b.lastError = err
			b.AddError(err)
		}
		return model
	}

	if cmd.View == nil {
		err := fmt.Errorf("[%s] has an unimplemented View func()", cmdPath)
		if b.lastError.Error() != err.Error() {
			logger.Log(logger.Error, "%s", err)
			b.lastError = err
			b.AddError(err)
		}
		return model
	}

	model = cmd.Run(model)
	errs := b.getErrors()
	if len(errs) > 0 {
		if b.lastError.Error() != errs[0].Error() {
			logger.Log(logger.Error, "%s", errs[0].Error())
		}
		b.lastError = errs[0]
	}

	return model
}

func (b Base[T]) getErrors() []error {
	return b.errors
}

func (b *Base[T]) clearErrors() {
	if len(b.errors) > 0 {
		b.lastError = fmt.Errorf("")
		b.errors = nil
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
