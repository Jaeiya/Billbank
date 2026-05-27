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
	columnStyle = Style.
			BorderBottom(true).
			Background(BgColor).
			BorderBackground(BgColor).
			BorderStyle(lipgloss.NormalBorder())
	rowStyle = Style.Background(BgColor)
)

type TableModel struct {
	table struct {
		entries    [][]TableEntry
		alignments []lipgloss.Position
		rowLen     int
	}
	header struct {
		values     []TableHeader
		alignments []lipgloss.Position
	}
	size struct {
		width, height int
	}
	selectedRow int
}

type TableOption func(*TableModel) error

type TableHeader struct {
	Name         string
	Width        int
	Foreground   color.Color
	Background   color.Color
	SelectedBack color.Color
}

type TableEntry struct {
	Text       string
	Foreground color.Color
	Background color.Color
}

func NewTable(
	headers []TableHeader,
	entries [][]TableEntry,
	opts ...TableOption,
) (TableModel, error) {
	t := TableModel{}
	t.table.rowLen = len(headers)

	if t.table.rowLen != len(entries[0]) {
		return t, errors.New(
			"table headers and data do not have the same length",
		)
	}

	t.table.rowLen = len(headers)
	t.table.entries = entries
	t.header.values = headers
	t.header.alignments = make([]lipgloss.Position, len(headers))
	t.table.alignments = make([]lipgloss.Position, len(headers))

	for _, o := range opts {
		err := o(&t)
		if err != nil {
			return t, err
		}
	}

	return t, nil
}

func WithDataAlignments(alignments []lipgloss.Position) TableOption {
	return func(tm *TableModel) error {
		if len(alignments) != len(tm.header.values) {
			return fmt.Errorf("table requires %d alignment positions", len(tm.header.values))
		}
		tm.table.alignments = alignments
		return nil
	}
}

func WithHeaderAlignments(alignments []lipgloss.Position) TableOption {
	return func(tm *TableModel) error {
		if len(alignments) != len(tm.header.values) {
			return fmt.Errorf("table requires %d alignment positions", len(tm.header.values))
		}
		tm.header.alignments = alignments
		return nil
	}
}

func (tm TableModel) Update(msg tea.Msg) (TableModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "j":
			if tm.selectedRow+1 == len(tm.table.entries) {
				return tm, nil
			}
			tm.selectedRow += 1
		case "k":
			if tm.selectedRow-1 < 0 {
				return tm, nil
			}
			tm.selectedRow -= 1
		case "h":
			tm.selectedRow = 0
		case "l":
			tm.selectedRow = len(tm.table.entries) - 1
		}
	}
	return tm, nil
}

func (tm TableModel) View() tea.View {
	return tea.NewView(
		Place(
			tm.size.width,
			tm.size.height,
			lipgloss.Center,
			lipgloss.Center,
			JoinVertical(
				lipgloss.Left,
				tm.ColumnView(),
				tm.RowView(),
			),
			lipgloss.WithWhitespaceStyle(
				lipgloss.NewStyle().Background(lipgloss.Color("#1E1E2E")),
			),
		),
	)
}

func (tm *TableModel) SetWidth(w int) {
	tm.size.width = w
}

func (tm *TableModel) SetHeight(h int) {
	tm.size.height = h
}

func (tm TableModel) SelectedRow() int {
	return tm.selectedRow
}

func (tm *TableModel) SetRow(rowIdx int, entries []TableEntry) error {
	if rowIdx >= len(tm.table.entries) || rowIdx < 0 {
		return errors.New("specified row index does not exist")
	}
	if len(entries) != tm.table.rowLen {
		return fmt.Errorf("too many or too few row entires")
	}
	tm.table.entries[rowIdx] = entries
	return nil
}

func (tm TableModel) ColumnView() string {
	cols := make([]string, len(tm.header.values))
	for i, h := range tm.header.values {
		if h.Width == 0 {
			h.Width = utils.RuneCount(h.Name)
		}

		h.Name = utils.TruncateStr(h.Name, h.Width)
		cols[i] = columnStyle.
			AlignHorizontal(tm.header.alignments[i]).
			Width(h.Width).
			Render(h.Name)
	}
	return JoinHorizontal(lipgloss.Left, cols...)
}

func (tm TableModel) RowView() string {
	rows := make([]string, len(tm.table.entries))
	entryBuf := make([]string, len(tm.table.entries[0]))

	for i, entries := range tm.table.entries {
		for k, entry := range entries {
			width := tm.header.values[k].Width
			if width == 0 {
				width = utils.RuneCount(entry.Text)
			}
			entry.Text = utils.TruncateStr(entry.Text, width)
			s := rowStyle.
				AlignHorizontal(tm.table.alignments[k]).
				Width(width)
			if i == tm.selectedRow {
				s = s.Background(Black)
			}
			s = s.Foreground(entry.Foreground)
			entryBuf[k] = s.Render(entry.Text)
		}
		rows[i] = JoinHorizontal(lipgloss.Left, entryBuf...)
	}
	return JoinVertical(lipgloss.Left, rows...)
}
