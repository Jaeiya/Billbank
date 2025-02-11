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
			{"this", "that", "other", "nil", "nilview"},
		}),
	}

	m.AddBranch([]cmd.Branch[skeletonModel]{
		// If you want to use aliases, you'll need to add
		// a separate branch with the same funcs. We have
		// two branches below that reference "this".
		{String: "t this", Fn: loadThis, ViewFn: thisView},
		{String: "test this", Fn: loadThis, ViewFn: thisView},

		{String: "t that", Fn: loadThat, ViewFn: thatView},

		// You can also set nil values
		{String: "t nil", Fn: nil, ViewFn: nil},
		{
			String: "t nilview",
			Fn:     func(sm skeletonModel) skeletonModel { return sm },
			ViewFn: nil,
		},
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
			logger.Log(logger.Info, "Skeleton", "hello from new command")
		}

		// Only watch certain keys on certain branches
		if m.IsActiveBranch("t this") || m.IsActiveBranch("test this") {
			if msg.String() == "ctrl+k" {
				m.thisCounter += 1
			}
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
