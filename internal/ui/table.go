package ui

import (
	"fmt"

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
	data struct {
		values     [][]string
		alignments []lipgloss.Position
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
	Name  string
	Width int
}

func NewTable(headers []TableHeader, data [][]string, opts ...TableOption) (TableModel, error) {
	t := TableModel{}
	t.data.values = data
	t.header.values = headers
	t.header.alignments = make([]lipgloss.Position, len(headers))
	t.data.alignments = make([]lipgloss.Position, len(headers))

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
		tm.data.alignments = alignments
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
			if tm.selectedRow+1 == len(tm.data.values) {
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
			tm.selectedRow = len(tm.data.values) - 1
		}
	}
	return tm, nil
}

func (tm TableModel) View() tea.View {
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

	rows := make([]string, len(tm.data.values))
	for i, data := range tm.data.values {
		row := make([]string, len(data))
		for k, val := range data {
			width := tm.header.values[k].Width
			if width == 0 {
				width = utils.RuneCount(val)
			}
			val = utils.TruncateStr(val, width)
			s := rowStyle.
				AlignHorizontal(tm.data.alignments[k]).
				Width(width)
			if i == tm.selectedRow {
				s = s.Background(Black)
			}
			row[k] = s.Render(val)
		}
		rows[i] = JoinHorizontal(lipgloss.Left, row...)
	}
	return tea.NewView(
		Place(
			tm.size.width,
			tm.size.height,
			lipgloss.Center,
			lipgloss.Center,
			JoinVertical(
				lipgloss.Left,
				JoinHorizontal(lipgloss.Left, cols...),
				JoinVertical(lipgloss.Left, rows...),
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
