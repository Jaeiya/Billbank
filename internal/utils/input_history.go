package utils

import (
	tea "charm.land/bubbletea/v2"
)

type InputHistory struct {
	cmds []string
	pos  int
}

func NewCmdHistory() *InputHistory {
	return &InputHistory{
		pos: -1,
	}
}

func (h *InputHistory) Add(v string) {
	if len(h.cmds) == 0 || h.cmds[0] != v {
		h.cmds = append([]string{v}, h.cmds...)
	}
	// Retain expected order when command is entered using history
	h.pos = -1
}

/*
Cycle moves either up or down through the history of commands
that have been added. Returns a string containing the command
text if it exists, and true if history cycles back to present.
*/
func (h *InputHistory) Cycle(msg tea.KeyMsg) (string, bool) {
	if len(h.cmds) == 0 {
		return "", false
	}

	switch msg.String() {
	case "up", "alt+k":
		if h.pos+1 < len(h.cmds) {
			h.pos++
		}

	case "down", "alt+j":
		if h.pos < 1 {
			h.pos = -1
			return "", true
		}
		h.pos--
	}

	return h.cmds[h.pos], false
}

func (h InputHistory) ListHistory() []string {
	newList := make([]string, len(h.cmds))
	copy(newList, h.cmds)
	return newList
}

func (h InputHistory) GetLen() int {
	return len(h.cmds)
}
