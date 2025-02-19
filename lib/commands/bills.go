package commands

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib/cmdmodel"
)

type billsModel struct {
	*cmdmodel.ModelBase[billsModel]
}

type billCmd = cmdmodel.BaseCommand[billsModel]

func NewBillsCmd() cmdmodel.Model {
	return cmdmodel.New(newBillsModel(
		"Bills",
		[]string{"bills"},
		[]billCmd{
			{
				Path: "",
				Run:  func(bm billsModel) billsModel { return bm },
				View: func(bm billsModel) string { return "bills command" },
			},
		}),
	)
}

func newBillsModel(name string, aliases []string, commands []billCmd) billsModel {
	return billsModel{
		ModelBase: cmdmodel.NewModelBase(cmdmodel.BaseCmdData[billsModel]{
			Name:     name,
			Aliases:  aliases,
			Commands: commands,
		}),
	}
}

func (m billsModel) Update(msg tea.Msg) (cmdmodel.Interface, tea.Cmd) {
	var teaCmd tea.Cmd
	var teaCmds []tea.Cmd

	//- DO NOT REMOVE; required for base model interaction
	m, teaCmd = m.ModelBase.Update(m, msg)
	teaCmds = append(teaCmds, teaCmd)
	//--------------------------------------------------//

	// switch msg := msg.(type) {
	// case tea.KeyMsg:
	// 	if msg.String() == "ctrl+k" {
	// 		m.thisCounter += 1
	// 	}
	// }

	return m, tea.Batch(teaCmds...)
}

func (m billsModel) View() string {
	// DO NOT REMOVE; required for base model interaction
	return m.ModelBase.View(m)
}

//##########################################
//     Custom Functions Go Below Here
//##########################################
