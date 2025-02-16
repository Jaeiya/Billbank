package commands

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib/cmdmodel"
)

func NewBillsCmd() cmdmodel.Model {
	//
	// Anything that needs to be initialized should be here
	// so that it can be passed to your model.
	//
	m := billsModel{
		ModelBase: cmdmodel.NewModelBase(cmdmodel.BaseData[billsModel]{
			Name:    "Bills",
			Aliases: []string{"bills", "b"},
			Commands: []cmdmodel.BaseCommand[billsModel]{
				{Path: "", Run: loadBills, View: viewBills},
			},
		}),
	}

	return cmdmodel.New(m)
}

type billsModel struct {
	*cmdmodel.ModelBase[billsModel]
}

func (m billsModel) Update(msg tea.Msg) (cmdmodel.Interface, tea.Cmd) {
	var teaCmd tea.Cmd
	var teaCmds []tea.Cmd

	m, teaCmd = m.ModelBase.Update(m, msg)
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
	return m.ModelBase.View(m)
}

func loadBills(m billsModel) billsModel {
	return m
}

func viewBills(m billsModel) string {
	return "Hello, " + m.GetCmdArg()
}
