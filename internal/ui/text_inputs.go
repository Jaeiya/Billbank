package ui

import (
	"charm.land/bubbles/v2/textinput"
)

func NewDefaultInput(width int) textinput.Model {
	m := textinput.New()
	m.Focus()
	m.SetVirtualCursor(true)
	s := m.Styles()
	s.Cursor.Blink = true
	s.Blurred.Text = s.Blurred.Text.Foreground(White)
	s.Focused.Placeholder = s.Focused.Placeholder.Foreground(Gray)
	s.Blurred.Placeholder = s.Blurred.Placeholder.Foreground(Gray)
	m.SetStyles(s)
	m.SetWidth(width)
	m.CharLimit = width
	return m
}
