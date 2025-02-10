
import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib/cmd"
	"github.com/jaeiya/billbank/lib/utils/logger"
)

func NewSkeletonCmd() cmd.Command {
	m := skeletonModel{
		BaseCmdModel: cmd.NewBaseModel[skeletonModel]([][]string{
			// Aliases
			{"t", "test"},
			// Branches
			{"this", "that", "other"},
		}),
	}

	m.AddBranch([]cmd.Branch[skeletonModel]{
		{String: "t this", Fn: loadThis, ViewFn: thisView, IsPollingKey: false},
		{String: "t that", Fn: loadThat, ViewFn: thatView, IsPollingKey: false},
		// You can also set a nil view
		{String: "t other", Fn: nil, ViewFn: nil, IsPollingKey: false},
	}...)

	return cmd.New(
		cmd.Config{
			Model:               m,
			InputValidationFunc: func(arg string) error { return nil },
			KeyValidationFunc:   func(key rune) bool { return false },
			HasArg:              false,
		},
	)
}

type skeletonModel struct {
	*cmd.BaseCmdModel[skeletonModel]
	thisCounter int
	thatCounter int
}

func (m skeletonModel) Update(msg tea.Msg) (cmd.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	m, cmd = m.BaseCmdModel.Update(m, msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+h" {
			logger.Log(logger.Info, "hello from new command")
		}

		if msg.String() == "ctrl+o" {
			m.thatCounter += 1
		}
	}

	return m, tea.Batch(cmds...)
}

func (m skeletonModel) View() string {
	return m.BaseCmdModel.View(m)
}

func loadThis(m skeletonModel) skeletonModel {
	return m
}

func thisView(m skeletonModel) string {
	return fmt.Sprintf(
		"This Counter goes up every time you enter the command: %d",
		m.thisCounter,
	)
}

func loadThat(m skeletonModel) skeletonModel {
	return m
}

func thatView(m skeletonModel) string {
	return fmt.Sprintf("This counter goes up every 300ms: %d", m.thatCounter)
}
