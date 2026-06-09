package ui

import (
	"errors"
	"fmt"
	"image/color"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/jaeiya/billbank/internal/utils"
)

var (
	headerStyle = Style.
			Bold(true).
			BorderBottom(true).
			Background(BgColor).
			BorderBackground(BgColor).
			BorderStyle(lipgloss.NormalBorder())

	rowStyle = Style.Background(BgColor)
)


type (
	TableHeader struct {
		Name      string
		Width     int
		Alignment lipgloss.Position
	}

	TableEntry struct {
		Text       string
		Foreground color.Color
	}

	TableData struct {
		Headers []TableHeader
		Entries [][]TableEntry
	}

	TableOption func(*TableModel) error

	TableModel struct {
		width, height   int
		headers         []TableHeader
		entries         [][]TableEntry
		entryAlignments []lipgloss.Position
		style           struct {
			header lipgloss.Style
			row    lipgloss.Style
		}
		selectedRow int
	}
)

func NewTable(d TableData, opts ...TableOption) (TableModel, error) {
	t := TableModel{}
	if len(d.Headers) == 0 {
		return t, errors.New("missing table headers")
	}

	if len(d.Entries) == 0 {
		return t, errors.New("missing table entries")
	}

	// This is a naive check; we assume subsequent entries match
	if len(d.Headers) != len(d.Entries[0]) {
		return t, errors.New(
			"table entry length should match table header length",
		)
	}

	t.entries = d.Entries
	t.headers = d.Headers
	t.entryAlignments = make([]lipgloss.Position, len(d.Headers))

	t.style.header = headerStyle
	t.style.row = rowStyle

	for _, o := range opts {
		err := o(&t)
		if err != nil {
			return t, err
		}
	}

	return t, nil
}

func WithColAlignments(alignments []lipgloss.Position) TableOption {
	return func(tm *TableModel) error {
		if len(alignments) != len(tm.headers) {
			return fmt.Errorf("table requires %d alignment positions", len(tm.headers))
		}
		tm.entryAlignments = alignments
		return nil
	}
}

func (table TableModel) Update(msg tea.Msg) (TableModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "j":
			if table.selectedRow+1 == len(table.entries) {
				return table, nil
			}
			table.selectedRow += 1
		case "k":
			if table.selectedRow-1 < 0 {
				return table, nil
			}
			table.selectedRow -= 1
		case "h":
			table.selectedRow = 0
		case "l":
			table.selectedRow = len(table.entries) - 1
		}
	}
	return table, nil
}

func (table TableModel) View() tea.View {
	return tea.NewView(
		JoinVertical(
			lipgloss.Left,
			table.headerView(),
			table.rowView(),
		),
	)
}

func (table *TableModel) SetWidth(w int) {
	table.width = w
}

func (table *TableModel) SetHeight(h int) {
	table.height = h
}

func (table TableModel) SelectedRow() int {
	return table.selectedRow
}

func (table *TableModel) SetRow(rowIdx int, entries []TableEntry) error {
	if rowIdx >= len(table.entries) || rowIdx < 0 {
		return errors.New("specified row index does not exist")
	}
	if len(entries) != len(table.headers) {
		return errors.New("too many or too few entries for row length")
	}
	table.entries[rowIdx] = entries
	return nil
}

func (table *TableModel) SetHeaderStyle(s lipgloss.Style) {
	table.style.header = s.Inherit(table.style.header)
}

func (table *TableModel) SetRowStyle(s lipgloss.Style) {
	table.style.row = s.Inherit(table.style.row)
}

func (table TableModel) headerView() string {
	headers := make([]string, len(table.headers))
	for i, h := range table.headers {
		if h.Width == 0 {
			h.Width = utils.RuneCount(h.Name)
		}

		h.Name = utils.TruncateStr(h.Name, h.Width)
		headers[i] = table.style.header.
			AlignHorizontal(table.headers[i].Alignment).
			Width(h.Width).
			Render(h.Name)
	}
	return JoinHorizontal(lipgloss.Left, headers...)
}

func (table TableModel) rowView() string {
	rows := make([]string, len(table.entries))
	entryBuf := make([]string, len(table.entries[0]))

	for i, entries := range table.entries {
		for k, entry := range entries {
			width := table.headers[k].Width
			if width == 0 {
				width = utils.RuneCount(entry.Text)
			}
			entry.Text = utils.TruncateStr(entry.Text, width)
			s := table.style.row.
				AlignHorizontal(table.entryAlignments[k]).
				Width(width)
			if i == table.selectedRow {
				s = s.Background(Black)
			}
			s = s.Foreground(entry.Foreground)
			entryBuf[k] = s.Render(entry.Text)
		}
		rows[i] = JoinHorizontal(lipgloss.Left, entryBuf...)
	}
	return JoinVertical(lipgloss.Left, rows...)
}
