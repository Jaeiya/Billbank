package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib/utils"
)

var cmdMap = map[string]func(DebugCmd) string{
	"history": getHistory,
	"log":     getLog,
	"slog":    getSLog,
}

func NewDebugCmd(h *utils.CmdHistory) Command {
	return NewCommand(
		CommandConfig{
			Command: Command{
				tree: [][]string{
					{"/"},
					listCommands(),
				},
				GetModel: func(status CommandStatus) CommandModelMsg {
					return DebugCmd{CommandStatus: status, history: h}
				},
				hasArg: false,
			},
		},
	)
}

type DebugCmd struct {
	CommandStatus
	history *utils.CmdHistory
}

func (DebugCmd) Init() tea.Cmd {
	return nil
}

func (x DebugCmd) Update(tea.Msg) (CommandModelMsg, tea.Cmd) {
	return x, nil
}

func (m DebugCmd) View() string {
	if exec, ok := cmdMap[m.CommandStr]; ok {
		return exec(m)
	}

	panic(fmt.Sprintf("missing view for '%s'", m.CommandStr))
}

func (DebugCmd) IsStatic() bool {
	return true
}

func getHistory(m DebugCmd) string {
	return m.history.View()
}

func getLog(DebugCmd) string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	path := filepath.Join(dir, "log.txt")
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err.Error()
	}
	return string(bytes)
}

func getSLog(DebugCmd) string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	path := filepath.Join(dir, "log.txt")
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err.Error()
	}

	str := strings.TrimSpace(string(bytes))
	lines := strings.Split(str, "\n")

	var sb strings.Builder
	for _, line := range lines {
		sb.WriteString(line[16:] + "\n")
	}
	return sb.String()
}

func listCommands() []string {
	list := make([]string, 0, len(cmdMap))
	for k := range cmdMap {
		list = append(list, k)
	}
	return list
}
