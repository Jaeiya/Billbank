
import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/jaeiya/billbank/internal/cmdcore"
	"github.com/jaeiya/billbank/internal/utils"
)

// Command model that shares all data between your commands.
// (rename to reflect the name of your new command)
type skeletonModel struct {
	cmdcore.ModelBase
	thisCounter int
	thatCounter int
	lastKey     string
	someInt     int
}

// NewSkeletonHandler demonstrates how to initialize a new
// command handler.
// (rename it to reflect the command name)
func NewSkeletonHandler() cmdcore.CommandHandler {
	// Initialize our model with the base command model
	// so we're not having to re-implement for each
	// command.
	model := skeletonModel{
		ModelBase: cmdcore.NewModelBase(),
	}

	commands := []cmdcore.Command[skeletonModel]{
		// Creates "skeleton" command which demonstrates how to create
		// what's called a 'default command' where the alias
		// can be executed as a command itself.
		//
		// 🟠 If you set the arg type to anything other than ArgNone,
		// then you can no longer add any other commands to the path.
		// This is because all subsequent words in the path would be
		// translated as args for the default command.
		cmdcore.NewCommand(cmdcore.CommandOptions[skeletonModel, cmdcore.NoArg]{
			Path:     "",
			RunFunc:  loadDummy,
			ViewFunc: loadDefaultView,
		}),

		// Creates a "skeleton a" command which increments a
		// counter when ctrl+k is pressed.
		cmdcore.NewCommand(cmdcore.CommandOptions[skeletonModel, cmdcore.NoArg]{
			Path:     "a",
			RunFunc:  loadDummy,
			ViewFunc: viewA,
		}),

		// Creates a "skeleton b" command which increments a
		// counter every time the command is executed.
		cmdcore.NewCommand(cmdcore.CommandOptions[skeletonModel, cmdcore.NoArg]{
			Path:     "b",
			RunFunc:  loadB,
			ViewFunc: viewB,
		}),

		// Creates "skeleton c" command which demonstrates how
		// to parse a required argument and receive it properly
		// with error handling. Optional arguments also require
		// a parse func.
		cmdcore.NewCommand(cmdcore.CommandOptions[skeletonModel, int]{
			Path:     "c",
			RunFunc:  loadC,
			ViewFunc: viewC,
			ArgType:  cmdcore.ArgRequired,
			ParseFunc: func(arg string) (int, error) {
				i, err := utils.ParseInt(arg)
				if err != nil {
					return 0, err
				}
				return i, nil
			},
		}),

		// Creates a "skeleton d" command which demonstrates how
		// a command can hide the input bar and immediately
		// receive model updates like key presses.
		cmdcore.NewCommand(cmdcore.CommandOptions[skeletonModel, cmdcore.NoArg]{
			Path:         "d",
			RunFunc:      loadDummy,
			ViewFunc:     viewD,
			CaptureInput: true,
		}),

		// Creates a "skeleton e" command which demonstrates how to get the
		// view port size.
		cmdcore.NewCommand(cmdcore.CommandOptions[skeletonModel, cmdcore.NoArg]{
			Path:     "e",
			RunFunc:  loadDummy,
			ViewFunc: viewE,
		}),
	}

	return cmdcore.NewCmdHandler(
		"skeleton",
		[]string{"skeleton"},
		commands,
		model,
	)
}

// Update is a required method to allow interaction with
// the TUI loop (via messages) and the command.
func (m skeletonModel) Update(msg tea.Msg) (cmdcore.CommandModel, tea.Cmd) {
	var teaCmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.IsWorkingPath("skeleton a") {
			if msg.String() == "ctrl+k" {
				m.thisCounter += 1
			}
		}
		m.lastKey = msg.String()
	}

	return m, tea.Batch(teaCmds...)
}

func loadDummy(m skeletonModel, arg *cmdcore.NoArg) (cmdcore.CommandModel, error) {
	return m, nil
}

func loadDefaultView(m skeletonModel) tea.View {
	return tea.NewView("default view")
}

func viewA(m skeletonModel) tea.View {
	return tea.NewView(fmt.Sprintf(
		`Hit ctrl+k to increment this counter: %d

Did it not work? That's because you'll need to hit the
Grave key (key below Esc key) to toggle the input. By
default, the command box takes input precedence.

Commands that need input from the user on load, can
use the CaptureInput flag, which will hide the
command box on command execution.
`,
		m.thisCounter,
	))
}

func loadB(m skeletonModel, arg *cmdcore.NoArg) (cmdcore.CommandModel, error) {
	m.thatCounter += 1
	return m, nil
}

func viewB(m skeletonModel) tea.View {
	text := `
Execute the command again to increment the counter: %d

When resizing the window, you will not see the
counter update, even though technically the
view is re-rendered. This is because window
resizes do not execute command logic.

This means that you'll need to place all
layout logic into the view. The reason for
this "feature" is because it prevents
expensive commands from lagging the UI
during window resizing.
`
	return tea.NewView(fmt.Sprintf(text, m.thatCounter))
}

func loadC(m skeletonModel, arg *int) (cmdcore.CommandModel, error) {
	m.someInt = *arg // do something with the arg
	return m, nil
}

func viewC(m skeletonModel) tea.View {
	return tea.NewView(fmt.Sprintf(`
You entered the number: %d

Validating arguments is very important and the error
messages returned to the user should always clue them
in on what they did wrong.
`, m.someInt))
}

func viewD(m skeletonModel) tea.View {
	return tea.NewView(fmt.Sprintf(`
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
`, m.lastKey))
}

func viewE(m skeletonModel) tea.View {
	v := tea.NewView("")
	w, h := m.ViewportSize()
	content := `
Width: %d, Height: %d

Resize the terminal to see the dimensions change in real time.
	`
	v.SetContent(fmt.Sprintf(content, w, h))
	return v
}
