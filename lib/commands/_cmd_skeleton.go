import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib/cmdmodel"
)

// Main model that contains all data for your new command.
// (rename to reflect the name of your new command)
type skeletonModel struct {
	*cmdmodel.Base[skeletonModel]
	thisCounter int
	thatCounter int
}

// Shortened type for convenience
// (rename it to reflect the command name)
type skeletonCmd = cmdmodel.Command[skeletonModel]

// Hook up all the commands to the command model
// (rename it to reflect the command name)
var skeletonCommands = []skeletonCmd{
	// Default command (cmd executed by itself)
	{Path: "", Run: loadDefaultCmd, View: loadDefaultView},
	// Creates "test this" command
	{Path: "this", Run: loadThis, View: thisView},
	// Creates "test that" command
	{Path: "that", Run: loadThat, View: thatView},
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
	case tea.KeyMsg:
		if msg.String() == "ctrl+k" {
			m.thisCounter += 1
		}
	}

	return m, tea.Batch(teaCmds...)
}

// DO NOT REMOVE or MODIFY; required for base model interaction
func (m skeletonModel) View() string {
	return m.Base.View(m)
}

//##########################################
//     Custom Functions Go Below Here
//##########################################

func loadDefaultCmd(m skeletonModel) skeletonModel {
	return m
}

func loadDefaultView(m skeletonModel) string {
	return "default view"
}

func loadThis(m skeletonModel) skeletonModel {
	return m
}

func thisView(m skeletonModel) string {
	return fmt.Sprintf(
		"Hit ctrl+k to increment this counter: %d",
		m.thisCounter,
	)
}

func loadThat(m skeletonModel) skeletonModel {
	m.thatCounter += 1
	return m
}

func thatView(m skeletonModel) string {
	text := `
Execute the command again to increment that counter: %d

If you resize the window, it will also increment,
this is because layout recalculations should happen
when the window is resized.
`
	return fmt.Sprintf(text, m.thatCounter)
}
