package ui

import (
	"errors"
	"fmt"
	"image/color"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/jaeiya/billbank/internal/utils"
)

var tableStyles = struct {
	header lipgloss.Style
	row    lipgloss.Style
}{
	header: Style.
		Bold(true).
		BorderBottom(true).
		Background(BgColor).
		BorderBackground(BgColor).
		BorderForeground(Gray).
		BorderStyle(lipgloss.NormalBorder()),

	row: Style.Background(BgColor),
}

type (
	TableEntry struct {
		Text       string
		Foreground color.Color
	}

	TableHeader struct {
		Name      string
		Width     int
		Alignment lipgloss.Position
	}

	TableData struct {
		Headers []TableHeader
		Entries [][]TableEntry
	}

	TableOption func(*TableModel) error

	TableStyles struct {
		Header lipgloss.Style
	}

	TableModel struct {
		width, height   int
		headers         []TableHeader
		entries         [][]TableEntry
		entryAlignments []lipgloss.Position
		style           TableStyles
		selectedRow     int
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

func (t TableModel) Update(msg tea.Msg) (TableModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "j":
			if t.selectedRow+1 == len(t.entries) {
				return t, nil
			}
			t.selectedRow += 1
		case "k":
			if t.selectedRow-1 < 0 {
				return t, nil
			}
			t.selectedRow -= 1
		case "h":
			t.selectedRow = 0
		case "l":
			t.selectedRow = len(t.entries) - 1
		}
	}
	return t, nil
}

func (t TableModel) View() tea.View {
	return tea.NewView(
		JoinVertical(
			lipgloss.Left,
			t.headerView(),
			t.rowView(),
		),
	)
}

func (t *TableModel) SetWidth(w int) {
	t.width = w
}

func (t *TableModel) SetHeight(h int) {
	t.height = h
}

func (t TableModel) SelectedRow() int {
	return t.selectedRow
}

func (t *TableModel) SetRow(rowIdx int, entries []TableEntry) error {
	if rowIdx >= len(t.entries) || rowIdx < 0 {
		return errors.New("specified row index does not exist")
	}
	if len(entries) != len(t.headers) {
		return errors.New("too many or too few entries for row length")
	}
	t.entries[rowIdx] = entries
	return nil
}

func (t TableModel) Style() TableStyles {
	return t.style
}

func (t *TableModel) SetStyle(s TableStyles) {
	t.style = s
}

func (t *TableModel) SelectRow(rowIdx int) error {
	if rowIdx > len(t.entries)-1 || rowIdx < 0 {
		return errors.New("cannot select row; index out of range")
	}
	t.selectedRow = rowIdx
	return nil
}

func (t TableModel) headerView() string {
	headers := make([]string, len(t.headers))
	for i, h := range t.headers {
		if h.Width == 0 {
			h.Width = utils.RuneCount(h.Name)
		}

		h.Name = utils.TruncateStr(h.Name, h.Width)
		header := t.style.Header.
			AlignHorizontal(t.headers[i].Alignment).
			Width(h.Width).
			Inherit(tableStyles.header).
			Render(h.Name)

		headers[i] = header
	}
	return JoinHorizontal(lipgloss.Left, headers...)
}

func (t TableModel) rowView() string {
	rows := make([]string, len(t.entries))
	entryBuf := make([]string, len(t.entries[0]))

	for i, entries := range t.entries {
		for k, entry := range entries {
			width := t.headers[k].Width
			if width == 0 {
				width = utils.RuneCount(entry.Text)
			}
			s := Style.AlignHorizontal(t.entryAlignments[k]).Width(width)
			if i == t.selectedRow {
				s = s.Background(Black)
			}
			s = s.Foreground(entry.Foreground)
			entryBuf[k] = s.Render(utils.TruncateStr(entry.Text, width))
		}
		rows[i] = tableStyles.row.Render(JoinHorizontal(lipgloss.Left, entryBuf...))
	}
	return JoinVertical(lipgloss.Left, rows...)
}
