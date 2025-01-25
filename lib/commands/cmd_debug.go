package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jaeiya/billbank/lib/utils"
)

var (
	infoLogStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#33B0FF"))
	attnLogStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFDE00"))
	errLogStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF82E9"))
	msgStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#96F1D4"))
)

var cmdMap = map[string]func(DebugCmd) string{
	"history": getHistory,
	"log":     getLog,
	"slog":    getSLog,
}

func NewDebugCmd(h *utils.InputHistory) Command {
	return NewCommand(
		CommandConfig{
			Command: Command{
				tree: [][]string{
					{"/"},
					listCommands(),
				},
				GetModel: func(status CommandStatus) CommandModelMsg {
					return DebugCmd{CommandStatus: status, history: h, state: &DebugCmdState{}}
				},
				hasArg: false,
			},
		},
	)
}

type DebugCmdState struct {
	lastSlogRender     string
	lastLineCount      int
	lastSlogRenderTook time.Duration
}

type DebugCmd struct {
	CommandStatus
	history *utils.InputHistory
	state   *DebugCmdState
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

	panic("missing view; use IsImplemented to check for this error")
}

func (m DebugCmd) IsImplemented() bool {
	_, ok := cmdMap[m.CommandStatus.CommandStr]
	return ok
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

func getSLog(m DebugCmd) string {
	now := time.Now()
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

	if len(lines) == m.state.lastLineCount {
		return fmt.Sprintf(
			"%s Took: %s",
			m.state.lastSlogRender,
			m.state.lastSlogRenderTook,
		)
	}

	m.state.lastLineCount = len(lines)

	var sb strings.Builder
	for _, line := range lines {
		parts := strings.Split(line, " ")
		tag := parts[3]
		words := parts[5:]

		var msg string
		if strings.Contains(words[0], ":") {
			words[1] = msgStyle.Render(strings.Join(words[1:], " "))
			msg = fmt.Sprintf("%s %s", words[0], words[1])
		} else {
			msg = msgStyle.Render(strings.Join(words, " "))
		}

		if tag == "[NFO]" {
			tag = infoLogStyle.Render(tag)
		} else if tag == "[ATN]" {
			tag = attnLogStyle.Render(tag)
			msg = attnLogStyle.Render(msg)
		} else if tag == "[ERR]" {
			tag = errLogStyle.Render(tag)
			msg = errLogStyle.Render(msg)
		}

		sb.WriteString(fmt.Sprintf("%s %s\n", tag, msg))
	}

	m.state.lastSlogRender = sb.String()
	m.state.lastSlogRenderTook = time.Since(now)
	return sb.String()
}

func listCommands() []string {
	list := make([]string, 0, len(cmdMap))
	for k := range cmdMap {
		list = append(list, k)
	}
	return list
}
