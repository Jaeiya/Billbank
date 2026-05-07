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
}

type commandModel struct {
	vpSize      struct{ w, h int }
	workingPath string
}

func NewModelBase() ModelBase {
	return &commandModel{}
}

func (cm commandModel) WorkingPath() string {
	return cm.workingPath
}

func (cm *commandModel) SetWorkingPath(path string) {
	cm.workingPath = path
}

func (cm commandModel) IsWorkingPath(path string) bool {
	return cm.workingPath == path
}

func (cm commandModel) ViewportSize() (w, h int) {
	return cm.vpSize.w, cm.vpSize.h
}

func (cm *commandModel) SetViewportSize(w, h int) {
	cm.vpSize.w = w
	cm.vpSize.h = h
}

func (cm commandModel) Reload(path string) tea.Cmd {
	if cm.IsWorkingPath(path) {
		return func() tea.Msg {
			return ExecCmdMsg{cm.vpSize.w, cm.vpSize.h}
		}
	}
	return nil
}
