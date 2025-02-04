package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jaeiya/billbank/lib"
)

type boxColor struct {
	bg     lipgloss.Color
	header lipgloss.Color
	msg    lipgloss.Color
}

var (
	errBoxColors = boxColor{
		bg:     lib.BgDarkColor,
		header: lib.FgErrColor,
		msg:    lib.FgWarnColor,
	}

	infoBoxColors = boxColor{
		bg:     lib.BgDarkColor,
		header: lib.FgSuccessColor,
		msg:    lib.FgColor,
	}

	boxSepChar = "━"

	boxHeader = lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Top).
			PaddingTop(1)

	boxMsgStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			PaddingBottom(1)
)

func NewErrorBox(title, msg string, width, height int) string {
	return newBox(width, height, title, msg, errBoxColors)
}

func NewInfoBox(title, msg string, width, height int) string {
	return newBox(width, height, title, msg, infoBoxColors)
}

func newBox(width, height int, title, msg string, color boxColor) string {
	centerWidth := width / 2
	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Left,
			boxHeader.Width(centerWidth).
				Background(color.bg).
				Foreground(color.header).
				Render(title),
			lipgloss.NewStyle().Background(lib.BgDarkColor).Foreground(lib.BgColor).
				Render(strings.Repeat(boxSepChar, centerWidth)),
			boxMsgStyle.Width(centerWidth).
				Background(color.bg).
				Foreground(color.msg).
				Render(msg),
		),
		lipgloss.WithWhitespaceBackground(lipgloss.Color("#1E1E2E")),
	)
}
