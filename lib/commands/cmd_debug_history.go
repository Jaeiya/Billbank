package commands

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib/utils"
)

type DebugHistoryCmd struct {
	history *utils.CmdHistory
}

func (DebugHistoryCmd) Init() tea.Cmd {
	return nil
}

func (x DebugHistoryCmd) Update(tea.Msg) (CommandModelMsg, tea.Cmd) {
	return x, nil
}

func (m DebugHistoryCmd) View() string {
	return m.history.View()
}

func (DebugHistoryCmd) IsStatic() bool {
	return true
}

func NewDebugHistoryCmd(h *utils.CmdHistory) Command {
	return NewCommand(
		CommandConfig{
			Command: Command{
				tree: [][]string{
					{"/x"},
					{"history"},
				},
				ModelMsg: DebugHistoryCmd{h},
				hasArg:   false,
			},
		},
	)
}
