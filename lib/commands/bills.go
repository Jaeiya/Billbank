package commands

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib/cmd"
)

func NewBillsCmd() cmd.Command {
	//
	// Anything that needs to be initialized should be here
	// so that it can be passed to your model.
	//
	m := billsModel{
		BaseModel: cmd.NewBaseModel[billsModel](cmd.Tree{
			Aliases: []string{"bills"},
			Branches: []cmd.CmdBranch{
				{Leaves: []string{}, HasArg: false},
			},
		}),
	}

	m.AddBranch([]cmd.Branch[billsModel]{}...)

	return cmd.New(
		cmd.Config{
			Model:               m,
			InputValidationFunc: func(arg string) error { return nil },
			KeyValidationFunc:   func(key rune) bool { return false },
			// You cannot mix arg & non-arg commands
			HasArg: false,
		},
	)
}

type billsModel struct {
	*cmd.BaseModel[billsModel]
}

func (m billsModel) Update(msg tea.Msg) (cmd.Model, tea.Cmd) {
	var teaCmd tea.Cmd
	var teaCmds []tea.Cmd

	m, teaCmd = m.BaseModel.Update(m, msg)
	teaCmds = append(teaCmds, teaCmd)

	// switch msg := msg.(type) {
	// case tea.KeyMsg:

	// 	// Only watch certain keys on certain branches
	// 	// if m.IsActiveBranch("t this") || m.IsActiveBranch("test this") {
	// 	// 	if msg.String() == "ctrl+k" {
	// 	// 		m.thisCounter += 1
	// 	// 	}
	// 	// }
	// }

	return m, tea.Batch(teaCmds...)
}

func (m billsModel) View() string {
	return m.BaseModel.View(m)
}
