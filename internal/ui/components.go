package ui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
	"github.com/jaeiya/billbank/internal/utils"
)

type BoxOptions struct {
	Content string
	Width   int
	Align   lipgloss.Position
}

func NewBox(opts BoxOptions) string {
	content := opts.Content
	runes := []rune(content)
	width := opts.Width
	const minWidth = 3

	if width <= minWidth {
		width = len(runes)
	}

	if width > minWidth && width < len(content) {
		runes = runes[:width-2]
		runes = append(runes, '.', '.')
	}

	content = string(runes)

	var sb strings.Builder
	borders := utils.GetBorderStyle(utils.SingleLine)

	// Compensates for spacing
	borderCount := width + 2

	// Compensates for vertical borders
	borderBytes := (borderCount + 4) * utf8.RuneLen(borders.Horizontal)

	sb.Grow(len(content) + (borderBytes * 2))

	sb.WriteRune(borders.TopLeft)
	sb.WriteString(utils.RepeatRune(borders.Horizontal, borderCount))
	sb.WriteRune(borders.TopRight)
	sb.WriteRune('\n')
	fmt.Fprintf(
		&sb,
		"%c %s %c\n",
		borders.Vertical,
		lipgloss.NewStyle().Width(width).AlignHorizontal(opts.Align).Render(content),
		borders.Vertical,
	)
	sb.WriteRune(borders.BottomLeft)
	sb.WriteString(utils.RepeatRune(borders.Horizontal, borderCount))
	sb.WriteRune(borders.BottomRight)
	return sb.String()
}

func NewBoxArray(width int, boxes ...string) string {
	return lipgloss.Place(
		width,
		3, // a box is always: border, text, border
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinHorizontal(lipgloss.Left, boxes...),
		lipgloss.WithWhitespaceStyle(
			lipgloss.NewStyle().Background(lipgloss.Color("#1E1E2E")),
		),
	)
}
