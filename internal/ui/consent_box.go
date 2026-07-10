package ui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type ConsentModel struct {
	msg     string
	buttons struct {
		yes Button
		no  Button
	}
	styles struct {
		box        lipgloss.Style
		btnWrapper lipgloss.Style
	}
}

func NewConsentBox(startOnYes bool) ConsentModel {
	m := ConsentModel{}

	const boxWidth = 45

	m.styles.box = Style.
		Width(boxWidth).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(DarkBorderColor).
		Foreground(Blue)

	m.styles.btnWrapper = Style.
		PaddingTop(1).
		Width(boxWidth - 4).
		Align(lipgloss.Right)

	btnYes := NewButton("Yes")
	btnNo := NewButton("No")

	if startOnYes {
		btnYes.Focus()
	} else {
		btnNo.Focus()
	}

	s := btnYes.Styles()
	s.Text.Focused = s.Text.Focused.Foreground(BrightGreen)
	s.Text.Blurred = s.Text.Blurred.Foreground(Gray)
	s.Selected = s.Selected.Foreground(BrightWhite)
	btnYes.SetStyle(s)
	btnNo.SetStyle(s)

	s = btnNo.Styles()
	s.Text.Focused = s.Text.Focused.Foreground(BrightRed)
	btnNo.SetStyle(s)

	m.buttons.yes = btnYes
	m.buttons.no = btnNo
	return m
}

func (m ConsentModel) Update(msg tea.Msg) (ConsentModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "h", "left":
			m.buttons.no.Blur()
			m.buttons.yes.Focus()
		case "l", "right":
			m.buttons.yes.Blur()
			m.buttons.no.Focus()
		}
	}

	return m, nil
}

func (m ConsentModel) View() string {
	btnView := JoinHorizontal(lipgloss.Left, m.buttons.yes.View()+" ", m.buttons.no.View())
	return m.styles.box.Render(
		JoinVertical(lipgloss.Left, m.msg, m.styles.btnWrapper.Render(btnView)),
	)
}

func (m *ConsentModel) SetMsg(msg string) {
	m.msg = msg
}
