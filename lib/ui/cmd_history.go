package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type CmdHistory struct {
	cmds []string
	pos  int
}

func NewCmdHistory() *CmdHistory {
	return &CmdHistory{
		pos: -1,
	}
}

func (h *CmdHistory) Enter(m CmdInputModel) {
	h.cmds = append([]string{m.CommandInput.Value()}, h.cmds...)
	// Retain expected order when command is entered using history
	h.pos = -1
}

func (h *CmdHistory) Cycle(m CmdInputModel, msg tea.KeyMsg) (CmdInputModel, tea.Cmd) {
	if len(h.cmds) == 0 {
		return m, nil
	}

	switch msg.String() {
	case "up", "alt+k":
		if h.pos+1 < len(h.cmds) {
			h.pos++
			m.CommandInput.SetValue(h.cmds[h.pos])
			m.CommandInput.CursorEnd()
		}

	case "down", "alt+j":
		if h.pos > 0 {
			h.pos--
			m.CommandInput.SetValue(h.cmds[h.pos])
			m.CommandInput.CursorEnd()
		} else {
			h.pos = -1
			m.CommandInput.Reset()
		}
	}

	return tryParseCmd(m, msg)
}
