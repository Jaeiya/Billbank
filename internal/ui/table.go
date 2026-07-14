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
	TableRow []TableColumn

	TableColumn struct {
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
		Rows    []TableRow
	}

	TableOption func(*TableModel) error

	TableStyles struct {
		Header lipgloss.Style
	}

	TableModel struct {
		width, height int
		headers       []TableHeader
		rows          []TableRow
		colAlignments []lipgloss.Position
		style         TableStyles
		selectedRow   int
	}
)

func NewTable(d TableData, opts ...TableOption) (TableModel, error) {
	t := TableModel{}
	if len(d.Headers) == 0 {
		return t, errors.New("missing table headers")
	}

	if len(d.Rows) == 0 {
		return t, errors.New("missing table rows")
	}

	// WARN: This is a naive check; we assume subsequent rows match
	if len(d.Headers) != len(d.Rows[0]) {
		return t, errors.New(
			"table row length should match table header length",
		)
	}

	t.rows = d.Rows
	t.headers = d.Headers
	t.colAlignments = make([]lipgloss.Position, len(d.Headers))

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
		tm.colAlignments = alignments
		return nil
	}
}

func (t TableModel) Update(msg tea.Msg) (TableModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "j":
			if t.selectedRow+1 == len(t.rows) {
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
			t.selectedRow = len(t.rows) - 1
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

func (t *TableModel) SetRow(rowIdx int, row TableRow) error {
	if rowIdx >= len(t.rows) || rowIdx < 0 {
		return errors.New("specified row index does not exist")
	}
	if len(row) != len(t.headers) {
		return errors.New("too many or too few columns for row length")
	}
	t.rows[rowIdx] = row
	return nil
}

func (t TableModel) Style() TableStyles {
	return t.style
}

func (t *TableModel) SetStyle(s TableStyles) {
	t.style = s
}

func (t *TableModel) SelectRow(rowIdx int) error {
	if rowIdx > len(t.rows)-1 || rowIdx < 0 {
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
	rows := make([]string, len(t.rows))
	colBuf := make([]string, len(t.rows[0]))

	for i, row := range t.rows {
		for k, col := range row {
			width := t.headers[k].Width
			if width == 0 {
				width = utils.RuneCount(col.Text)
			}
			s := Style.AlignHorizontal(t.colAlignments[k]).Width(width)
			if i == t.selectedRow {
				s = s.Background(Black)
			}
			s = s.Foreground(col.Foreground)
			colBuf[k] = s.Render(utils.TruncateStr(col.Text, width))
		}
		rows[i] = tableStyles.row.Render(JoinHorizontal(lipgloss.Left, colBuf...))
	}
	return JoinVertical(lipgloss.Left, rows...)
}
