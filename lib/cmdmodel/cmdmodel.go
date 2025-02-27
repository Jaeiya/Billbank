package cmdmodel

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	MsgMissingArgFuncErr = `
If a command path can accept an argument, then it should also validate
that argument. It's recommended to validate the arg inside this
function because it can notify the user on error.`

	MsgIsCmdItselfErr = `
Models with default commands (commands executed just by their alias),
that can take arguments, do not support extra commands. For instance,
if your model has a "view" alias that takes an argument for the kind
of view to display:

[view 1] or [view 2]

You're limited to just a single command path, the default path. You
cannot then create more commands that take a more specific view type
like so:

[view details] or [view list]

The above second-order commands will be ignored as if they don't exist.
Either setup your model to accept args directly or commands, but not
both. You can also setup your model to execute a default command without
args, allowing you to add extra commands.`

	MsgUsingAliasInCmdPathErr = `
Command-paths do not need to include the alias of their parent model.

If you're trying to setup a default command (a command executed by its
aliases alone) then just add a command with an empty string for its path
like so:

{ Path: "", Run: loadCmd, View: loadView }

Where loadCmd and loadView are functions that take your command model
as an argument. This will allow the execution of the model aliases,
as if they were commands themselves.

Be aware though, that if you set its ArgType to optional or required,
you'll no longer be able to add any more commands to that model.`

	MsgDuplicateAliasErr = `
Check to make sure you don't already have a command with
that alias. You may also have accidentally added the
command more than once.`

	MsgInvalidHomeCmdPathErr = `
Make sure you've entered the entire command path, including the alias.
You also cannot set a home path that requires arguments.

Double check the command paths of the command model you're trying to
access and make sure the path exists.`
)

var (
	ErrNotCommand = fmt.Errorf("unrecognized command")
	// Command is valid, but entered incorrectly
	ErrIncompleteCmd   = fmt.Errorf("improper command entry")
	ErrMissingArgument = fmt.Errorf("missing argument")
	ErrEmptyCommand    = fmt.Errorf("empty command")
	// Command is missing an execution path
	ErrUnimplementedCmd = fmt.Errorf("unimplemented command")
)

type Interface interface {
	Update(tea.Msg) (Interface, tea.Cmd)
	View() string
	GetName() string
	GetAliases() []string
	GetId() int
	GetCmdPaths() []string
	ParseCommand(string) Status
	GetStatus() Status
	SetStatus(Status)
	IsInitialized() bool
}

type Status struct {
	Suggestions  []string
	Path         string
	Arg          any
	Error        error
	IsCommand    bool
	CaptureInput bool
}

type (
	HomeMsg struct{}

	StatusBarMsg struct {
		String   string
		Severity StatusSeverity
	}

	ViewportSizeMsg struct {
		Width  int
		Height int
	}

	UpdateCmdMsg struct {
		Model Interface
	}
)
