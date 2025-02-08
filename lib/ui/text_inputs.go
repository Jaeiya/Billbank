package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

const blinkSpeed = time.Millisecond * 500

func NewDefaultInput(width int) textinput.Model {
	m := textinput.New()
	m.Focus()
	m.Cursor.Style = m.Cursor.Style.Foreground(lipgloss.Color("#00FFA2"))
	m.Cursor.BlinkSpeed = blinkSpeed
	m.Width = width
	m.CharLimit = width
	return m
}
