package commands

import (
	tea "charm.land/bubbletea/v2"
	"github.com/jaeiya/billbank/internal/cmdcore"
	"github.com/jaeiya/billbank/internal/db/sqlite"
)

type billsModel struct {
	cmdcore.ModelBase
	db *sqlite.SqliteDb
}

func NewBillsHandler(db *sqlite.SqliteDb) cmdcore.CommandHandler {
	model := billsModel{
		ModelBase: cmdcore.NewModelBase(),
	}
	h := cmdcore.NewCmdHandler(
		"bills",
		[]string{"bills"},
		[]cmdcore.Command[billsModel]{
			cmdcore.NewCommand(cmdcore.CommandOptions[billsModel, cmdcore.NoArg]{
				Path:     "",
				RunFunc:  loadBills,
				ViewFunc: viewBills,
			}),
		},
		model,
	)
	return h
}

func (m billsModel) Update(msg tea.Msg) (cmdcore.CommandModel, tea.Cmd) {
	return m, nil
}

func loadBills(m billsModel, arg *cmdcore.NoArg) (cmdcore.CommandModel, error) {
	return m, nil
}

func viewBills(m billsModel) tea.View {
	return tea.NewView("hello world")
}
