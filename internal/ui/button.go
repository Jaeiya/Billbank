package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type ButtonMsg struct {
	Name string
}

type Button struct {
	isFocused bool
	text      string
	styles    ButtonStyles
}

type ButtonStyles struct {
	Text struct {
		Focused lipgloss.Style
		Blurred lipgloss.Style
	}
	Selected lipgloss.Style
}

func NewButton(text string) Button {
	b := Button{
		text: text,
	}
	return b
}

func (b Button) Update(msg tea.Msg) (Button, tea.Cmd) {
	return b, nil
}

func (b Button) View() string {
	var textBuilder strings.Builder
	textBuilder.Grow(len(b.text) + 2)

	bracketStyle := b.styles.Selected.Inline(true)
	text := b.styles.Text.Blurred.Inline(true).Render(b.text)

	if b.isFocused {
		text = b.styles.Text.Focused.Inline(true).Render(b.text)
		textBuilder.WriteString(bracketStyle.Render("["))
		textBuilder.WriteString(text)
		textBuilder.WriteString(bracketStyle.Render("]"))
	} else {
		textBuilder.WriteString(" ")
		textBuilder.WriteString(text)
		textBuilder.WriteString(" ")
	}

	return Style.Render(textBuilder.String())
}

func (b *Button) Focus() {
	b.isFocused = true
}

func (b Button) Focused() bool {
	return b.isFocused
}

func (b *Button) Blur() {
	b.isFocused = false
}

func (b Button) Styles() ButtonStyles {
	return b.styles
}

func (b *Button) SetStyle(s ButtonStyles) {
	b.styles = s
}

func (b Button) Press() tea.Cmd {
	return func() tea.Msg {
		return ButtonMsg{b.text}
	}
}
