package components

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

const blinkSpeed = time.Millisecond * 400

func NewDefaultInput(width int) textinput.Model {
	m := textinput.New()
	m.Focus()
	m.Cursor.Style = m.Cursor.Style.Foreground(lipgloss.Color("#00FFA2"))
	m.Cursor.BlinkSpeed = blinkSpeed
	m.Width = width
	m.CharLimit = width
	return m
}

func NewCommanderInput() textinput.Model {
	m := textinput.New()
	m.Prompt = ""
	m.PlaceholderStyle = m.PlaceholderStyle.Foreground(lipgloss.Color("#00FFA2"))
	m.CompletionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFA2"))
	m.Focus()
	m.Cursor.Style = m.Cursor.Style.Foreground(lipgloss.Color("#00FFA2"))
	m.PromptStyle = m.Cursor.Style.Foreground(lipgloss.Color("#00FFA2"))
	m.Cursor.BlinkSpeed = blinkSpeed
	m.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFF"))
	m.Prompt = "> "
	m.ShowSuggestions = true
	return m
}
