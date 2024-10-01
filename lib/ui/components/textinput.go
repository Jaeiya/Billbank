package components

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

func NewTextInput() textinput.Model {
	m := textinput.New()
	m.Prompt = ""
	m.PlaceholderStyle = m.PlaceholderStyle.Foreground(lipgloss.Color("#00FFA2"))
	m.Focus()
	m.Cursor.Style = m.Cursor.Style.Foreground(lipgloss.Color("#00FFA2"))
	m.Cursor.BlinkSpeed = time.Millisecond * 450
	m.CharLimit = 156
	m.Width = 50
	m.Prompt = ""
	m.ShowSuggestions = true
	return m
}
