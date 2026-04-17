import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/jaeiya/billbank/lib/cmdmodel"
	"github.com/jaeiya/billbank/lib/utils"
)

// Main model that contains all data for your new command.
// (rename to reflect the name of your new command)
type skeletonModel struct {
	*cmdmodel.Base[skeletonModel]
	thisCounter int
	thatCounter int
	lastKey     string
	cInt        int
}

// Shortened type for convenience
// (rename it to reflect the command name)
type skeletonCmd = cmdmodel.Command[skeletonModel]

// Hook up all the commands to the command model
// (rename it to reflect the command name)
var skeletonCommands = []skeletonCmd{
	// Creates "test" command which demonstrates how to create
	// what's called a 'default command' where the alias
	// can be executed by itself. If you add the arguments
	// flag, then it can receive arguments.
	{Path: "", Run: loadDummy, View: loadDefaultView},

	// Creates a "test a" command which increments a
	// counter when ctrl+k is pressed.
	{Path: "a", Run: loadDummy, View: viewA},

	// Creates a "test b" command which increments a
	// counter every time the command is executed.
	{Path: "b", Run: loadB, View: viewB},

	// Creates "test c" command which demonstrates how
	// to parse a required argument and receive it properly
	// with error handling. Optional arguments also require
	// a parse method.
	{
		Path:    "c",
		Run:     loadC,
		View:    viewC,
		ArgType: cmdmodel.ArgRequired,
		// All arguments require this func, which converts
		// the user input into a valid value. If there's
		// an error, it should return a user-readable
		// message.
		ParseArg: func(arg string) (any, error) {
			i, err := utils.ParseInt(arg)
			if err != nil {
				if strings.Contains(err.Error(), "out of range") {
					return nil, fmt.Errorf("number is too large")
				}
				return nil, fmt.Errorf("'%s' is not an integer", arg)
			}
			return i, nil
		},
	},

	// Creates a "test d" command which demonstrates how
	// the CaptureInput flag works.
	{Path: "d", Run: loadDummy, View: viewD, CaptureInput: true},
}

// newSkeletonModel demonstrates how to initialize a new
// command model. All initialization code for your model
// should go in here.
// (rename it to reflect the command name)
func NewSkeletonCmd() skeletonModel {
	//
	// Any initialization code should go here
	//
	return skeletonModel{
		Base: cmdmodel.NewBaseModel(cmdmodel.CommandData[skeletonModel]{
			// Command name (should be unique)
			Name: "Skeleton",
			// Prefixes all command paths ("test <cmd_path>")
			Aliases:  []string{"test"},
			Commands: skeletonCommands,
		}),
	}
}

func (m skeletonModel) Update(msg tea.Msg) (cmdmodel.Interface, tea.Cmd) {
	var teaCmd tea.Cmd
	var teaCmds []tea.Cmd

	//- DO NOT REMOVE or MODIFY; required for base model interaction
	m, teaCmd = m.Base.Update(m, msg)
	teaCmds = append(teaCmds, teaCmd)
	//--------------------------------------------------//

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+k" {
			m.thisCounter += 1
		}

		m.lastKey = msg.String()
	}

	return m, tea.Batch(teaCmds...)
}

// DO NOT REMOVE or MODIFY; required for base model interaction
func (m skeletonModel) View() string {
	return m.Base.View(m)
}

// ##########################################
//
//	Custom Functions Go Below Here
//
// ##########################################
func loadDummy(m skeletonModel) skeletonModel {
	return m
}

func loadDefaultView(m skeletonModel) string {
	return "default view"
}

func viewA(m skeletonModel) string {
	return fmt.Sprintf(
		`Hit ctrl+k to increment this counter: %d

Did it not work? That's because you'll need to hit the
Grave key (key below Esc key) to toggle the input. By
default, the command box takes input precedence.

Commands that need input from the user on load, can
use the CaptureInput flag, which will hide the
command box on command execution.
`,
		m.thisCounter,
	)
}

func loadB(m skeletonModel) skeletonModel {
	m.thatCounter += 1
	return m
}

func viewB(m skeletonModel) string {
	text := `
Execute the command again to increment the counter: %d

When resizing the window, you will not see the
counter update, even though technically the
view is re-rendered. This is because window
resizes do not re-execute command logic.

This means that you'll need to place all
layout logic into the view. The reason for
this "feature" is because it prevents
expensive commands from lagging the UI
during window resizing.
`
	return fmt.Sprintf(text, m.thatCounter)
}

func loadC(m skeletonModel) skeletonModel {
	// This func is very important because it will propagate
	// configuration errors if it finds any. This is the
	// only reliable way to retrieve command arguments.
	myInt, isValid := cmdmodel.GetArgAs[int](m.Base)
	if !isValid {
		return m
	}

	m.cInt = myInt // do something with the arg
	return m
}

func viewC(m skeletonModel) string {
	return fmt.Sprintf(`
You entered the number: %d

Validating arguments is very important and the error
messages returned to the user should always clue them
in on what they did wrong.
`, m.cInt)
}

func viewD(m skeletonModel) string {
	return fmt.Sprintf(`
Last key entered: %s

This view captures the input immediately
without you having to hit the grave key first.

Try typing some keys and see what happens!

The grave key (key under Esc key) toggles the
command input box at the bottom of the screen.
When it's hidden, all key input is transferred
to the view.

This means if the command input is active, you
won't see the keys change above, because the
command input will be intercepting all of
them.
`, m.lastKey)
}
