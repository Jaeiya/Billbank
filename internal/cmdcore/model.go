package cmdcore

import tea "charm.land/bubbletea/v2"

type CommandModel interface {
	ModelBase
	Update(msg tea.Msg) (CommandModel, tea.Cmd)
}

type ModelBase interface {
	WorkingPath() string
	SetWorkingPath(path string)
	IsWorkingPath(path string) bool
	ViewportSize() (w, h int)
	SetViewportSize(w, h int)
	// Loads the specified command path if it
	// is already active.
	Reload(path string) tea.Cmd
	SendErr(err error) tea.Cmd
}

type commandModel struct {
	vpSize      struct{ w, h int }
	workingPath string
}

func NewModelBase() ModelBase {
	return &commandModel{}
}

func (m commandModel) WorkingPath() string {
	return m.workingPath
}

func (m *commandModel) SetWorkingPath(path string) {
	m.workingPath = path
}

func (m commandModel) IsWorkingPath(path string) bool {
	return m.workingPath == path
}

func (m commandModel) ViewportSize() (w, h int) {
	return m.vpSize.w, m.vpSize.h
}

func (m *commandModel) SetViewportSize(w, h int) {
	m.vpSize.w = w
	m.vpSize.h = h
}

func (m commandModel) Reload(path string) tea.Cmd {
	if m.IsWorkingPath(path) {
		return func() tea.Msg {
			return ExecCmdMsg{m.vpSize.w, m.vpSize.h}
		}
	}
	return nil
}

func (m commandModel) SendErr(err error) tea.Cmd {
	return func() tea.Msg {
		return FatalCmdErrMsg(err)
	}
}
