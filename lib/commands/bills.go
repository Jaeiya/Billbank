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
		BaseModel: cmd.NewBaseModel(cmd.Tree{
			Aliases: []string{"bills"},
			Branches: []cmd.Branch{
				{Leaves: []string{}, HasArg: false},
			},
		}, []cmd.BranchCommand[billsModel]{
			{Path: "bills", Run: loadBills, View: viewBills},
		}),
	}

	return cmd.New(
		cmd.Config{
			Name:  "Bills",
			Model: m,
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

func loadBills(m billsModel) billsModel {
	return m
}

func viewBills(m billsModel) string {
	return "hello"
}
