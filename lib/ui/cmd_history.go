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

func (h *CmdHistory) Add(v string) {
	h.cmds = append([]string{v}, h.cmds...)
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
		}

	case "down", "alt+j":
		if h.pos < 1 {
			h.pos = -1
			m.CommandInput.Reset()
			m.lastCmd = ParsedCmd{}
			return m, nil
		}
		h.pos--
	}

	m.CommandInput.SetValue(h.cmds[h.pos])
	m.CommandInput.CursorEnd()

	return tryParseCmd(m, tea.KeyMsg{})
}
