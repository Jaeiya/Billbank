package commands

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/internal/cmdmodel"
	"github.com/jaeiya/billbank/internal/db/sqlite"
)

type billsModel struct {
	*cmdmodel.Base[billsModel]
	db *sqlite.SqliteDb
}

type billCmd = cmdmodel.Command[billsModel]

var billCommands = []billCmd{
	{Path: "", Run: loadBills, View: viewBills},
}

func NewBillsCmd(db *sqlite.SqliteDb) billsModel {
	return billsModel{
		Base: cmdmodel.NewBaseModel(cmdmodel.CommandData[billsModel]{
			Name:     "Bills",
			Aliases:  []string{"bills"},
			Commands: billCommands,
		}),
		db: db,
	}
}

func (m billsModel) Update(msg tea.Msg) (cmdmodel.Interface, tea.Cmd) {
	var teaCmd tea.Cmd
	var teaCmds []tea.Cmd

	//- DO NOT REMOVE or MODIFY; required for base model interaction
	m, teaCmd = m.Base.Update(m, msg)
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
	return m.Base.View(m)
}

//##########################################
//     Custom Functions Go Below Here
//##########################################

func loadBills(m billsModel) billsModel {
	return m
}

func viewBills(m billsModel) string {
	return "hello world"
}
