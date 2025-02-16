import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib/cmd"
)

func NewSkeletonCmd() cmd.Command {
	//
	// Anything that needs to be initialized should be here
	// so that it can be passed to your model.
	//
	m := skeletonModel{}

	return cmd.New(
		cmd.Config{
			Model: m,
		},
	)
}

type skeletonModel struct {
	*cmd.BaseModel[skeletonModel]
	thisCounter int
	thatCounter int
}

func (m skeletonModel) Update(msg tea.Msg) (cmd.Model, tea.Cmd) {
	var teaCmd tea.Cmd
	var teaCmds []tea.Cmd

	m, teaCmd = m.BaseModel.Update(m, msg)
	teaCmds = append(teaCmds, teaCmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:

	}

	return m, tea.Batch(teaCmds...)
}

func (m skeletonModel) View() string {
	return m.BaseModel.View(m)
}

func loadThis(m skeletonModel) skeletonModel {
	return m
}

func thisView(m skeletonModel) string {
	return fmt.Sprintf(
		"Hit ctrl+k to increment the counter: %d",
		m.thisCounter,
	)
}

func loadThat(m skeletonModel) skeletonModel {
	m.thatCounter += 1
	return m
}

func thatView(m skeletonModel) string {
	return fmt.Sprintf("Execute the command again to increment the counter: %d", m.thatCounter)
}
