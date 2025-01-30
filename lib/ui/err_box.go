package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	errBgColor     = bgDarkColor
	errHeaderColor = fgErrColor
	errMsgColor    = lipgloss.Color("#eee")
	errSeparatorFg = bgColor
	errSepChar     = "━"

	baseErrStyle = lipgloss.NewStyle().Background(errBgColor)

	errHeader = baseErrStyle.
			Align(lipgloss.Center, lipgloss.Top).
			Foreground(errHeaderColor).
			PaddingTop(1)

	errSeparator = baseErrStyle.
			Foreground(errSeparatorFg)

	errMsg = baseErrStyle.
		Foreground(errMsgColor).
		PaddingLeft(1).
		PaddingRight(1).
		PaddingBottom(1)
)

func NewErrorMsg(title string, msg string, termWidth int, termHeight int) string {
	centerWidth := termWidth / 2
	return lipgloss.Place(
		termWidth,
		termHeight,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Left,
			errHeader.Width(centerWidth).Render(title),
			errSeparator.Render(strings.Repeat(errSepChar, centerWidth)),
			errMsg.Width(centerWidth).Render(msg),
		),
		lipgloss.WithWhitespaceBackground(lipgloss.Color("#1E1E2E")),
	)
}
