package commands

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib/cmdmodel"
)

type billsModel struct {
	*cmdmodel.BaseModel[billsModel]
}

type billCmd = cmdmodel.BaseCommand[billsModel]

func NewBillsCmd() cmdmodel.Model {
	return cmdmodel.New(newBillsModel(
		"Bills",
		[]string{"bills"},
		[]billCmd{
			{Path: "", Run: loadBills, View: viewBills},
		}),
	)
}

func newBillsModel(name string, aliases []string, commands []billCmd) billsModel {
	return billsModel{
		BaseModel: cmdmodel.NewModelBase(cmdmodel.BaseCmdData[billsModel]{
			Name:     name,
			Aliases:  aliases,
			Commands: commands,
		}),
	}
}

func (m billsModel) Update(msg tea.Msg) (cmdmodel.Interface, tea.Cmd) {
	var teaCmd tea.Cmd
	var teaCmds []tea.Cmd

	//- DO NOT REMOVE or MODIFY; required for base model interaction
	m, teaCmd = m.BaseModel.Update(m, msg)
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

// DO NOT REMOVE or MODIFY; required for base model interaction
func (m billsModel) View() string {
	return m.BaseModel.View(m)
}

//##########################################
//     Custom Functions Go Below Here
//##########################################

func loadBills(m billsModel) billsModel {
	return m
}

func viewBills(m billsModel) string {
	return ""
}
