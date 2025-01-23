package utils

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var viewStyle = lipgloss.NewStyle().Width(30)

type CmdHistory struct {
	cmds          []string
	pos           int
	lastView      string
	lastViewedLen int
}

func NewCmdHistory() *CmdHistory {
	return &CmdHistory{
		pos: -1,
	}
}

func (h *CmdHistory) Add(v string) {
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
func (h *CmdHistory) Cycle(msg tea.KeyMsg) (string, bool) {
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

func (h *CmdHistory) View() string {
	if h.lastViewedLen == len(h.cmds) {
		return h.lastView
	}
	var sb strings.Builder
	for i, cmd := range h.cmds {
		if i == 0 {
			sb.WriteString(cmd)
			continue
		}
		sb.WriteString(fmt.Sprintf("\n%s", cmd))
	}
	h.lastViewedLen = len(h.cmds)
	h.lastView = viewStyle.Render(sb.String())
	return h.lastView
}
