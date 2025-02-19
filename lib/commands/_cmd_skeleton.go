import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib/cmdmodel"
)

// Main model that contains all data for your new command.
// (rename to reflect the name of your new command)
type skeletonModel struct {
	*cmdmodel.ModelBase[skeletonModel]
	thisCounter int
	thatCounter int
}

// Shorten type for convenience
// (rename it to reflect the command name)
type skeletonCmd = cmdmodel.BaseCommand[skeletonModel]

// Creates a new command using the base model
// (rename it to reflect the command name)
func NewSkeletonCmd() cmdmodel.Model {
	//
	// Do NOT put initialization code here
	//
	return cmdmodel.New(
		newSkeletonModel(
			"Skeleton",
			[]string{"test"},
			[]skeletonCmd{
				{Path: "", Run: loadDefaultCmd, View: loadDefaultView},
				{Path: "this", Run: loadThis, View: thisView},
				{Path: "that", Run: loadThat, View: thatView},
			}),
	)
}

// newSkeletonModel demonstrates how to initialize a new
// command model. All initialization code for your model
// should go in here.
// (rename it to reflect the command name)
func newSkeletonModel(name string, aliases []string, commands []skeletonCmd) skeletonModel {
	return skeletonModel{
		ModelBase: cmdmodel.NewModelBase(cmdmodel.BaseCmdData[skeletonModel]{
			Name:     name,
			Aliases:  aliases,
			Commands: commands,
		}),
	}
}

func (m skeletonModel) Update(msg tea.Msg) (cmdmodel.Interface, tea.Cmd) {
	var teaCmd tea.Cmd
	var teaCmds []tea.Cmd

	//- DO NOT REMOVE; required for base model interaction
	m, teaCmd = m.ModelBase.Update(m, msg)
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

func (m skeletonModel) View() string {
	// DO NOT REMOVE; required for base model interaction
	return m.ModelBase.View(m)
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
