package utils

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHistoryConstructor(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)
	h := NewCmdHistory()
	r.NotNil(h, "expect new history struct")
	a.Equal(-1, h.pos, "expect initial position to be -1")
	a.Len(h.cmds, 0, "expect empty command list")
}

func TestHistoryAdd(t *testing.T) {
	a := assert.New(t)
	h := NewCmdHistory()

	// Test adding first command
	h.Add("command1")
	a.Len(h.cmds, 1, "expected 1 command to be added")
	a.Equal("command1", h.cmds[0], "expected first value to be 'command1'")
	a.Equal(-1, h.pos, "expected position to be reset to -1")

	// Test adding second command
	h.Add("command2")
	a.Len(h.cmds, 2, "expected 2nd command to be added")
	a.Equal("command2", h.cmds[0], "expect 'command2' to be in 1st position")
	a.Equal("command1", h.cmds[1], "expect 'command1' to be in 2nd position")

	// Test adding duplicate command
	h.pos = 1 // Set position to something other than -1
	h.Add("command2")
	a.Len(h.cmds, 2, "expected same length of commands when adding duplicate")
	a.Equal(-1, h.pos, "expected position to reset to -1")
}

func TestHistoryCycle(t *testing.T) {
	h := NewCmdHistory()
	h.Add("command3")
	h.Add("command2")
	h.Add("command1")

	tests := []struct {
		should          string
		key             string
		expectedCmd     string
		expectedPresent bool
		expectedPos     int
	}{
		{"cycle up once", "up", "command1", false, 0},
		{"cycle up a 2nd time", "up", "command2", false, 1},
		{"cycle up a 3rd time", "up", "command3", false, 2},
		{"not cycle beyond end", "up", "command3", false, 2}, // No change when at end
		{"cycle down once", "down", "command2", false, 1},
		{"cycle down twice", "down", "command1", false, 0},
		{"cycle back to current", "down", "", true, -1},
		{"not cycle beyond current", "down", "", true, -1}, // No change when at beginning
		{"cycle up with Alt+k", "alt+k", "command1", false, 0},
		{"cycle down with Alt+j", "alt+j", "", true, -1},
	}

	for _, tt := range tests {
		t.Run("should "+tt.should, func(t *testing.T) {
			a := assert.New(t)
			cmd, present := h.Cycle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})
			a.Equal(tt.expectedCmd, cmd)
			a.Equal(tt.expectedPresent, present)
			a.Equal(tt.expectedPos, h.pos)
		})
	}
}

func TestHistoryCycleEmpty(t *testing.T) {
	h := NewCmdHistory()
	cmd, present := h.Cycle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("up")})
	a := assert.New(t)
	a.Equal("", cmd, "should have empty history")
	a.False(present, "should indicate missing history")
}

func TestHistoryList(t *testing.T) {
	a := assert.New(t)
	h := NewCmdHistory()
	h.Add("command3")
	h.Add("command2")
	h.Add("command1")

	list := h.ListHistory()
	a.Len(list, 3)

	// Verify ordering
	expected := []string{"command1", "command2", "command3"}
	a.Equal(expected, list)

	// Verify it's a copy (modify original)
	h.Add("command4")
	a.Len(list, 3, "should not modify returned history list")

	// Verify it's a copy (modify copy)
	list[0] = "modified"
	newList := h.ListHistory()
	a.NotEqual("modified", newList[0], "should return a copy of history list")
}

func TestHistoryLength(t *testing.T) {
	a := assert.New(t)
	h := NewCmdHistory()

	a.Equal(0, h.GetLen(), "expected 0 length for new history")

	h.Add("command1")
	a.Equal(1, h.GetLen(), "expect length 1 after adding command")

	h.Add("command2")
	a.Equal(2, h.GetLen(), "expect length 2 after adding 2nd command")

	h.Add("command2")
	a.Equal(2, h.GetLen(), "expect same length after adding duplicate command")
}
