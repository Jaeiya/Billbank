package commands

import (
	"errors"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
	if m.state == Setup {
		m.form, cmd = m.form.Update(msg)
		cmds = append(cmds, cmd)
	}

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

		m.form = newBillForm(m)
	}

	return m, nil
}

func viewBills(m billsModel) tea.View {
	w, _ := m.ViewportSize()
	// m.table.SetWidth(w)
	// m.table.SetHeight(h)
	if m.state == Setup {
		return tea.NewView(
			ui.Style.Width(w).MarginTop(2).Align(lipgloss.Center).Render(m.form.View()),
		)
	}
	return tea.NewView("")
}

func newBillForm(m billsModel) ui.Form {
	return ui.NewForm(
		"New Bill",
		ui.NewFormField().
			Title("Name").
			Description("The name of the bill.\n\nThis should be a concise name like: Bank of America, Property Tax, Health insurance, etc...").
			InputType(ui.AlphaNumInput),

		ui.NewFormField().
			Title("Type").
			Description("The type or category this bill falls into:\n\n"+viewBillTypes()).
			InputType(ui.AlphaInput).
			Suggestions(defaultBillTypes[:]...).
			Validator(func(s string, defaultValidator func() error, args ...string) error {
				for _, t := range defaultBillTypes {
					if s == t {
						return nil
					}
				}
				return errors.New("invalid bill type; pick from the above options")
			}),

		ui.NewFormField().
			Title("Amount").
			Description("The amount that the bill costs.").
			InputType(ui.CurrencyInput),

		ui.NewFormField().
			Title("Period").
			Description("The recurring time period that this bill needs to be paid:\n\n"+viewPeriods()).
			Suggestions(sqlite.PeriodStrings[:]...).
			Validator(func(s string, defaultValidator func() error, args ...string) error {
				for _, p := range sqlite.PeriodStrings {
					if s == p {
						return nil
					}
				}
				return errors.New("invalid period; pick from the above options")
			}),

		ui.NewFormField().
			Title("Due Date").
			LinkTo("Period").
			DescriptionDyn(func(args ...string) string {
				period := args[3]
				switch sqlite.Period(period) {
				case sqlite.YEARLY:
					return "Enter the last date this bill was paid, in the format: yyyy/mm/dd"
				case sqlite.Monthly:
					return "Enter the day of the month this bill is paid."
				case sqlite.BiMonthly:
					return "Enter the month and day this bill is paid."
				}
				return ""
			}),
	)
}

func viewPeriods() string {
	var leftSb strings.Builder
	leftSb.Grow(6 * len(sqlite.PeriodStrings))

	var rightSb strings.Builder
	rightSb.Grow(6 * len(sqlite.PeriodStrings))

	for i, p := range sqlite.PeriodStrings {
		if (i+1)%2 == 0 {
			rightSb.WriteString("  - ")
			rightSb.WriteString(p)
			rightSb.WriteRune('\n')
		} else {
			leftSb.WriteString("- ")
			leftSb.WriteString(p)
			leftSb.WriteRune('\n')
		}
	}

	return strings.TrimSpace(
		lipgloss.JoinHorizontal(lipgloss.Top, leftSb.String(), rightSb.String()),
	)
}

func viewBillTypes() string {
	var leftSb strings.Builder
	leftSb.Grow(8 * len(defaultBillTypes) / 2)

	var rightSb strings.Builder
	rightSb.Grow(8 * len(defaultBillTypes) / 2)

	for i, t := range defaultBillTypes {
		if (i+1)%2 == 0 {
			rightSb.WriteString("  - ")
			rightSb.WriteString(t)
			rightSb.WriteString("\n")
		} else {
			leftSb.WriteString("- ")
			leftSb.WriteString(t)
			leftSb.WriteString("\n")
		}
	}

	return strings.TrimSpace(
		ui.JoinHorizontal(lipgloss.Center, leftSb.String(), rightSb.String()),
	)
}
