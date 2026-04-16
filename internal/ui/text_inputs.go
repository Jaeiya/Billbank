package ui

import (
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const blinkSpeed = time.Millisecond * 500

func NewDefaultInput(width int) textinput.Model {
	m := textinput.New()
	m.Focus()
	c := m.Cursor()
	c.Color = lipgloss.Color("#00FFA2")
	c.Blink = true
	c.Shape = tea.CursorBar
	m.SetWidth(width)
	m.CharLimit = width
	return m
}
