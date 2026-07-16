package commands

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/jaeiya/billbank/internal"
	"github.com/jaeiya/billbank/internal/cmdcore"
	"github.com/jaeiya/billbank/internal/db/sqlite"
	"github.com/jaeiya/billbank/internal/ui"
	"github.com/jaeiya/billbank/internal/utils"
)

type BillCmdState int

const (
	Normal BillCmdState = iota
	Setup
	FinishSetup
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
	db        *sqlite.SqliteDb
	table     ui.TableModel
	form      ui.Form
	consent   ui.ConsentModel
	billNames *[]string
	state     BillCmdState
}

func NewBillsHandler(db *sqlite.SqliteDb) cmdcore.CommandHandler {
	model := billsModel{
		ModelBase: cmdcore.NewModelBase(),
		db:        db,
		consent:   ui.NewConsentBox(true),
		billNames: &[]string{},
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
	case ui.FormSavedMsg:
		if len(msg) != 5 {
			return m, m.SendErr(fmt.Errorf("fatal::expected form with 5 fields, but got %d", len(msg)))
		}
		m.consent.SetMsg("Would you like to add any more bills?")
		m.state = FinishSetup
		err := m.saveBill(msg)
		if err != nil {
			return m, m.SendErr(err)
		}

	case ui.ConsentMsg:
		switch m.state {
		// Does the user want to add more bills?
		case FinishSetup:
			if msg.Yes {
				cmd = m.form.Reset()
				m.state = Setup
			} else {
				m.state = Normal
				err := m.createBillsTable()
				if err != nil {
					return m, m.SendErr(err)
				}
			}
		}
	}

	cmds = append(cmds, cmd)

	switch m.state {
	case Normal:
		m.table, cmd = m.table.Update(msg)
	case Setup:
		m.form, cmd = m.form.Update(msg)
	case FinishSetup:
		m.consent, cmd = m.consent.Update(msg)

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
		typeRecords := make([]sqlite.BillTypeRecord, 0, len(sqlite.BillTypeMap))

		for typeName, id := range sqlite.BillTypeMap {
			typeRecords = append(typeRecords, sqlite.BillTypeRecord{ID: id, Name: typeName})
		}

		if err = m.db.CreateBillTypes(typeRecords); err != nil {
			return m, err
		}

		now := time.Now()
		if _, err = m.db.CreateMonth(now.Year(), now.Month()); err != nil {
			return m, err
		}

		m.form = m.createBillForm()
	}

	return m, nil
}

func viewBills(m billsModel) tea.View {
	w, _ := m.ViewportSize()

	switch m.state {
	case Normal:
		now := time.Now()
		return tea.NewView(
			ui.Style.
				MarginTop(2).
				Width(w).
				Align(lipgloss.Center).
				Render(
					ui.JoinVertical(
						lipgloss.Center,
						ui.ToAsciiFont(
							now.Format("Jan")+" "+strconv.Itoa(now.Year()),
							ui.FutureSmooth,
						),
						"",
						m.table.View().Content,
					),
				),
		)

	case Setup:
		return tea.NewView(
			ui.Style.Width(w).MarginTop(2).Align(lipgloss.Center).Render(m.form.View()),
		)

	case FinishSetup:
		return tea.NewView(
			ui.Style.Width(w).PaddingTop(5).Align(lipgloss.Center).Render(m.consent.View()),
		)
	}

	return tea.NewView("")
}

func (m *billsModel) saveBill(data []string) error {
	name, typeStr, amount, periodStr, dueDateStr := data[0], data[1], data[2], data[3], data[4]
	_ = dueDateStr

	*m.billNames = append(*m.billNames, name)

	typeID, exists := sqlite.BillTypeMap[typeStr]
	if !exists {
		return fmt.Errorf("fatal::missing bill type: '%s'", typeStr)
	}

	now := time.Now()
	period := sqlite.Period(periodStr)
	var dueDate internal.Date

	switch period {
	case sqlite.Yearly, sqlite.BiYearly, sqlite.TriYearly:
		dateParts := strings.Split(dueDateStr, "/")
		y, err := utils.ParseInt(dateParts[0])
		if err != nil {
			return fmt.Errorf("failed to parse year: %w", err)
		}

		m, err := utils.ParseInt(dateParts[1])
		if err != nil {
			return fmt.Errorf("failed to parse month: %w", err)
		}

		d, err := utils.ParseInt(dateParts[2])
		if err != nil {
			return fmt.Errorf("failed to parse day: %w", err)
		}

		dueDate, err = internal.NewDate(y, time.Month(m), d)
		if err != nil {
			return fmt.Errorf("failed to create due date: %w", err)
		}

	case sqlite.Monthly, sqlite.BiMonthly:
		dateParts := strings.Split(dueDateStr, "/")
		m, err := utils.ParseInt(dateParts[0])
		if err != nil {
			return fmt.Errorf("failed to parse month: %w", err)
		}

		d, err := utils.ParseInt(dateParts[1])
		if err != nil {
			return fmt.Errorf("failed to parse day: %w", err)
		}

		dueDate, err = internal.NewDate(now.Year(), time.Month(m), d)
		if err != nil {
			return fmt.Errorf("failed to create due date: %w", err)
		}

	case sqlite.Weekly, sqlite.BiWeekly:
		return fmt.Errorf("failed to parse due date: unsupported period [%s]", periodStr)
	}

	err := m.db.AddNewBills([]sqlite.BillRecord{
		{
			Name:    name,
			TypeID:  typeID,
			Amount:  internal.NewCurrency(amount, internal.USD),
			Period:  sqlite.Period(periodStr),
			Status:  sqlite.Pending,
			DueDate: dueDate,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to save bill: %w", err)
	}

	return nil
}

func (m *billsModel) createBillsTable() error {
	var err error

	records, err := m.db.QueryBills(sqlite.QueryMap{})
	if err != nil {
		return err
	}

	rows := make([]ui.TableRow, len(records))

	for i, r := range records {
		billType := sqlite.FindBillType(r.TypeID)
		ddTime := r.DueDate.GetTime()
		dueDate := ""
		switch r.Period {
		case sqlite.Yearly, sqlite.BiYearly, sqlite.TriYearly:
			dueDate = fmt.Sprintf("%02d/%02d", ddTime.Month(), ddTime.Day())
		case sqlite.Monthly, sqlite.BiMonthly:
			dueDate = fmt.Sprintf("%02d", ddTime.Day())
		}

		rows[i] = []ui.TableColumn{
			{Text: r.Name},
			{Text: r.Amount.String()},
			{Text: dueDate},
			{Text: string(r.Period)},
			{Text: billType},
		}
	}

	m.table, err = ui.NewTable(
		ui.TableData{
			Headers: []ui.TableHeader{
				{Name: "Name", Width: 20, Alignment: lipgloss.Left},
				{Name: "Amount", Width: 10, Alignment: lipgloss.Center},
				{Name: "Due Date", Width: 13, Alignment: lipgloss.Center},
				{Name: "Period", Width: 9, Alignment: lipgloss.Left},
				{Name: "Type", Width: 15, Alignment: lipgloss.Left},
			},
			Rows: rows,
		},
		ui.WithColAlignments([]lipgloss.Position{
			lipgloss.Left,
			lipgloss.Right,
			lipgloss.Center,
			lipgloss.Left,
			lipgloss.Left,
		}),
	)
	if err != nil {
		return err
	}
	return nil
}

func (m *billsModel) createBillForm() ui.Form {
	return ui.NewForm(
		"New Bill",
		ui.NewFormField().
			Title("Name").
			Description("The name of the bill.\n\nThis should be a concise name like: Bank of America, Property Tax, Health insurance, etc...").
			Placeholder("Rent").
			InputType(ui.AlphaNumInput).
			Validator(func(s string, defaultValidator func() error, args ...string) error {
				if utils.RuneCount(s) < 3 {
					return errors.New("invalid name; bill name must be more than 3 characters")
				}
				if slices.Contains(*m.billNames, s) {
					return errors.New("bill already exists; bill name must be unique")
				}
				return nil
			}),

		ui.NewFormField().
			Title("Type").
			Description("The type or category this bill falls into:\n\n"+viewBillTypes()).
			InputType(ui.AlphaInput).
			Placeholder("Utility").
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
			Placeholder("12.34").
			InputType(ui.CurrencyInput),

		ui.NewFormField().
			Title("Period").
			Description("The recurring time period that this bill needs to be paid:\n\n"+viewPeriods()).
			Placeholder("yearly").
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
			DescriptionDyn(func(args ...string) string {
				period := args[3]
				switch sqlite.Period(period) {
				case sqlite.Yearly, sqlite.BiYearly, sqlite.TriYearly:
					return "Enter the last date this bill was paid, in the format: yyyy/mm/dd"
				case sqlite.Monthly, sqlite.BiMonthly:
					return "Enter the Month & Day this bill is due.\n\nSupported formats:\n- 5/8\n- 05/08"
				case sqlite.Weekly, sqlite.BiWeekly:
					return "Unsupported Period"
				}
				return ""
			}).
			Placeholder("7/15").
			LinkTo("Period").
			Validator(func(s string, defaultValidator func() error, args ...string) error {
				period := args[0]
				switch sqlite.Period(period) {
				case sqlite.Yearly, sqlite.BiYearly:
					return validateYearlyBill(s)

				case sqlite.Monthly, sqlite.BiMonthly:
					return validateMonthlyBill(s)

				case sqlite.Weekly, sqlite.BiWeekly:
					return errors.New("weekly and bi-weekly are not implemented")

				default:
					return errors.New("fatal::missing due date validator for this option")
				}
			}),
	)
}

func validateMonthlyBill(s string) error {
	if strings.Count(s, "/") != 1 {
		return errors.New("invalid date: must have 1 slash (/) between month and day")
	}
	dateParts := strings.Split(s, "/")
	m := dateParts[0]
	d := dateParts[1]

	monthInt, err := utils.ParseInt(m)
	if err != nil {
		return fmt.Errorf("invalid date: %w", err)
	}

	if monthInt < 1 || monthInt > 12 {
		return errors.New("invalid month: must be a value in the range 1 - 12")
	}

	dayInt, err := utils.ParseInt(d)
	if err != nil {
		return fmt.Errorf("invalid day: %w", err)
	}

	if dayInt < 1 || dayInt > 31 {
		return errors.New("invalid day: must be a value in the range 1 - 31")
	}
	return nil
}

func validateYearlyBill(s string) error {
	if len(s) != 10 {
		return errors.New("date too short/long; required format: yyyy/mm/dd")
	}

	if strings.Count(s, "/") != 2 {
		return errors.New(
			"invalid date; must contain a '/' before and after the month and no more",
		)
	}

	dateParts := strings.Split(s, "/")
	y := dateParts[0]
	m := dateParts[1]
	d := dateParts[2]

	if len(y) != 4 {
		return errors.New(
			"invalid year; year should be 4 numeric digits like '2025'",
		)
	}

	yearInt, err := utils.ParseInt(y)
	if err != nil {
		return fmt.Errorf("invalid year: %w", err)
	}

	t := time.Now()

	if yearInt > t.Year() {
		return errors.New("invalid year: enter a past paid year")
	}

	if t.Year()-yearInt > 3 {
		return errors.New("invalid year: last paid year must be within 3 years")
	}

	monthInt, err := utils.ParseInt(m)
	if err != nil {
		return fmt.Errorf("invalid month: %w", err)
	}

	if monthInt < 1 || monthInt > 12 {
		return errors.New("invalid month: must be a value in the range 1 - 12")
	}

	dayInt, err := utils.ParseInt(d)
	if err != nil {
		return fmt.Errorf("invalid day: %w", err)
	}

	if dayInt < 1 || dayInt > 31 {
		return errors.New("invalid day: must be a value in the range 1 - 31")
	}

	return nil
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
