package commands

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/jaeiya/billbank/internal/cmdcore"
	"github.com/jaeiya/billbank/internal/db/sqlite"
	"github.com/jaeiya/billbank/internal/ui"
)

type BillCmdState int

const (
	Normal BillCmdState = iota
	Setup
)

var defaultBillTypes = [...]string{
	"credit card",
	"donation",
	"education",
	"entertainment",
	"family",
	"food",
	"health",
	"housing",
	"insurance",
	"loan",
	"maintenance",
	"medical",
	"subscription",
	"tax",
	"transit",
	"utility",
}

type billsModel struct {
	cmdcore.ModelBase
	db       *sqlite.SqliteDb
	table    ui.TableModel
	form     ui.Form
	state    BillCmdState
	initForm bool
	tableLen int
}

func NewBillsHandler(db *sqlite.SqliteDb) cmdcore.CommandHandler {
	model := billsModel{
		ModelBase: cmdcore.NewModelBase(),
		db:        db,
	}

	h := cmdcore.NewCmdHandler(
		"bills",
		[]string{"bills"},
		[]cmdcore.Command[billsModel]{
			cmdcore.NewCommand(cmdcore.CommandOptions[billsModel, cmdcore.NoArg]{
				Path:         "",
				RunFunc:      loadBills,
				ViewFunc:     viewBills,
				CaptureInput: true,
			}),
		},
		model,
	)

	return h
}

func (m billsModel) Update(msg tea.Msg) (cmdcore.CommandModel, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		// Exits application
		if msg.String() == "esc" {
			return m, tea.Quit
		}
	}

	cmds = append(cmds, cmd)

	return m, tea.Sequence(cmds...)
}

func loadBills(m billsModel, arg *cmdcore.NoArg) (cmdcore.CommandModel, error) {
	count, err := m.db.QueryBillsCount()
	if err != nil {
		panic(err)
	}

	if count == 0 {
		m.state = Setup

		if err = m.db.CreateBillTypes(defaultBillTypes[:]); err != nil {
			return m, err
		}

		now := time.Now()
		if _, err = m.db.CreateMonth(now.Year(), now.Month()); err != nil {
			return m, err
		}

	}

	return m, nil
}

func viewBills(m billsModel) tea.View {
	// w, h := m.ViewportSize()
	// m.table.SetWidth(w)
	// m.table.SetHeight(h)
	return tea.NewView("")
}
