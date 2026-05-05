package commands

import (
	tea "charm.land/bubbletea/v2"
	"github.com/jaeiya/billbank/internal/cmdcore"
	"github.com/jaeiya/billbank/internal/db/sqlite"
)

type billsModel struct {
	db          *sqlite.SqliteDb
	workingPath string
	vpSize      struct{ w, h int }
}

func NewBillsHandler(db *sqlite.SqliteDb) cmdcore.CommandHandler {
	model := billsModel{}
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

func (m billsModel) GetViewportSize() (w, h int) {
	return m.vpSize.w, m.vpSize.h
}

func (m billsModel) IsWorkingPath(path string) bool {
	return m.workingPath == path
}

func (m billsModel) SetWorkingPath(path string) cmdcore.CommandModel {
	m.workingPath = path
	return m
}

func (m billsModel) SetViewportSize(w, h int) cmdcore.CommandModel {
	m.vpSize.w = w
	m.vpSize.h = h
	return m
}

func loadBills(m billsModel, arg *cmdcore.NoArg) (cmdcore.CommandModel, error) {
	return m, nil
}

func viewBills(m billsModel) tea.View {
	return tea.NewView("hello world")
}
