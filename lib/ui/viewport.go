package ui

import tea "github.com/charmbracelet/bubbletea"

type CurrentCmd struct {
	Model tea.Model
	Cmd   tea.Cmd
}

type ViewPort struct {
	Commander  CmdInputModel
	CurrentCmd *CurrentCmd
	status     string
}

func (vp ViewPort) Init() tea.Cmd {
	return nil
}

func (vp ViewPort) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.(type) {
	case TestMsg:
		vp.status = "I executed because of a command!!"
	}
	vp.Commander, cmd = vp.Commander.Update(msg)
	return vp, cmd
}

func (vp ViewPort) View() string {
	if vp.status != "" {
		return vp.status
	}
	return vp.Commander.View()
}
