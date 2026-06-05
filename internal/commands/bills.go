package commands

import (
	"fmt"
	"math/rand/v2"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/jaeiya/billbank/internal/cmdcore"
	"github.com/jaeiya/billbank/internal/db/sqlite"
	"github.com/jaeiya/billbank/internal/ui"
	"github.com/jaeiya/billbank/internal/utils"
)

type billsModel struct {
	cmdcore.ModelBase
	db       *sqlite.SqliteDb
	table    ui.TableModel
	tableLen int
}

func NewBillsHandler(db *sqlite.SqliteDb) cmdcore.CommandHandler {
	model := billsModel{
		ModelBase: cmdcore.NewModelBase(),
	}

	rows := make([][]ui.TableEntry, 10)
	for i := range rows {
		rows[i] = createRow()
	}

	const (
		nameLen     = 20
		amountLen   = 11
		dueDayLen   = 10
		statusLen   = 10
		intervalLen = 11
		methodLen   = 15
	)

	model.tableLen = nameLen + amountLen + dueDayLen + statusLen + intervalLen + methodLen

	var err error
	model.table, err = ui.NewTable(
		[]ui.TableHeader{
			{Name: "Name", Width: nameLen},
			{Name: "Amount", Width: amountLen},
			{Name: "Due", Width: dueDayLen},
			{Name: "Status", Width: statusLen},
			{Name: "Interval", Width: intervalLen},
			{Name: "Payment Method", Width: methodLen},
		},
		rows,
		ui.WithDataAlignments([]lipgloss.Position{
			lipgloss.Left, lipgloss.Right, lipgloss.Center, lipgloss.Left, lipgloss.Center, lipgloss.Center,
		}),
		ui.WithHeaderAlignments([]lipgloss.Position{
			lipgloss.Left, lipgloss.Right, lipgloss.Center, lipgloss.Left, lipgloss.Center, lipgloss.Center,
		}),
	)
	if err != nil {
		panic(err)
	}

	model.table.SetHeaderStyle(
		ui.Style.Foreground(ui.RealWhite).Bold(false),
	)

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
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "esc" {
			return m, tea.Quit
		}

		if msg.String() == "enter" {
			idx := m.table.SelectedRow()
			err := m.table.SetRow(idx, createRow())
			if err != nil {
				panic(err)
			}
		}
	}
	m.table, _ = m.table.Update(msg)
	return m, nil
}

func loadBills(m billsModel, arg *cmdcore.NoArg) (cmdcore.CommandModel, error) {
	return m, nil
}

func viewBills(m billsModel) tea.View {
	w, h := m.ViewportSize()
	m.table.SetWidth(w)
	m.table.SetHeight(h)

	total := "1234.56"
	totalOffset := strings.Repeat(" ", 20+11-utils.RuneCount(total)) + total

	tableTotal := ui.JoinHorizontal(
		lipgloss.Left,
		ui.Style.Width(m.tableLen).
			BorderTop(true).
			Background(ui.BgColor).
			BorderBackground(ui.BgColor).
			BorderStyle(lipgloss.NormalBorder()).
			Foreground(ui.BrightGreen).
			Render(totalOffset),
	)

	header := ui.Style.
		PaddingTop(2).
		PaddingBottom(3).
		Background(ui.BgColor).
		Width(m.tableLen).
		Align(lipgloss.Center).
		Render(ui.ToAsciiFont("Bills", ui.BigMoney))

	return tea.NewView(
		ui.Place(
			w, h,
			lipgloss.Center, lipgloss.Top,
			ui.JoinVertical(lipgloss.Left, header, m.table.View().Content, tableTotal),
			lipgloss.WithWhitespaceStyle(lipgloss.NewStyle().Background(ui.BgColor)),
		),
	)
}

func createRow() []ui.TableEntry {
	names := [13]string{
		"Bank of America",
		"Netflix",
		"Amazon",
		"Cell Phone",
		"Mortgage",
		"Electric",
		"Water",
		"Sears",
		"Capital One",
		"Spectrum Internet",
		"Spotify",
		"Health Insurance",
		"Groceries",
	}

	highPrice := max(rand.IntN(1500), 30)
	lowPrice := max(30, rand.IntN(100))

	price := highPrice
	if rand.IntN(100) < 50 {
		price = lowPrice
	}
	cents := rand.IntN(99)

	paid := "Paid"
	paidColor := ui.BrightGreen
	if rand.IntN(100) < 30 {
		paid = "Missed"
		paidColor = ui.BrightRed
	} else if rand.IntN(100) < 50 {
		paid = "Pending"
		paidColor = ui.Gray
	}

	return []ui.TableEntry{
		{Text: names[rand.IntN(len(names))]},
		{Text: fmt.Sprintf("$%d.%02d", price, cents), Foreground: ui.Green},
		{Text: fmt.Sprintf("%02d", rand.IntN(31)+1)},
		{Text: paid, Foreground: paidColor},
		{Text: "monthly", Foreground: ui.Gray},
		{Text: "Credit Card", Foreground: ui.Gray},
	}
}
