package ui

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

type boxColor struct {
	bg     color.Color
	header color.Color
	msg    color.Color
}

var (
	errBoxColors = boxColor{
		bg:     BgDarkColor,
		header: FgErrColor,
		msg:    FgWarnColor,
	}

	infoBoxColors = boxColor{
		bg:     BgDarkColor,
		header: FgSuccessColor,
		msg:    FgColor,
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
	maxWidth := 60
	centerWidth := min(width/2, maxWidth)
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
			lipgloss.NewStyle().Background(BgDarkColor).Foreground(BgColor).
				Render(strings.Repeat(boxSepChar, centerWidth)),
			boxMsgStyle.Width(centerWidth).
				Background(color.bg).
				Foreground(color.msg).
				Render(msg),
		),
		lipgloss.WithWhitespaceStyle(
			lipgloss.NewStyle().Background(lipgloss.Color("#1E1E2E")).Padding(1, 2),
		),
	)
}
