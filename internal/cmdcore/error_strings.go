package cmdcore

const (
	MsgIsCmdItselfErr = `
Command models with default commands (commands executed just through their
models alias), that can take arguments, do not support extra commands.
For instance, if your model has a "view" alias that takes an argument
for the kind of view to display:

[view 1] or [view 2]

You're limited to just a single command path, the default path. You
cannot then create more commands that take a more specific view type
like so:

[view details] or [view list]

The above second-order commands will be ignored as if they don't exist.
Either setup your model to accept args or command paths, but not both.
You can also setup your model to execute a default command without args,
which still allows you to add extra command paths.`

	MsgUsingAliasInCmdPathErr = `
Command-paths do not need to include the alias of their parent command
model.

If you're trying to setup a default command (a command executed by its
parent model aliases) then just add a command with an empty string for
its path like so:

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

	MsgFatalMissingCmd = `
This means that the normal command parsing has been circumvented or
modified in a way that has made it brittle. For all intents and
purposes, this should NEVER happen.
`
)
