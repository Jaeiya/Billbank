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
	histStyle    = lipgloss.NewStyle().Width(30)
)

var cmdMap = map[string]func(DebugCmd) string{
	"history": getInputHistory,
	"log":     getLog,
	"slog":    getSLog,
}

func NewDebugCmd(h *utils.InputHistory) Command {
	dc := DebugCmd{
		history: h,
		state:   &DebugCmdState{slog: &DebugSlog{}, inputHistory: &DebugInputHistory{}},
	}

	cmds := make([]string, 0, len(cmdMap))
	for k := range cmdMap {
		cmds = append(cmds, k)
	}

	return NewCommand(
		CommandConfig{
			Command: Command{
				tree: [][]string{
					{"/"},
					cmds,
				},
				GetModel: func(status CommandStatus) CommandModelMsg {
					dc.CommandStatus = status
					return dc
				},
				hasArg: false,
			},
		},
	)
}

type DebugCmdState struct {
	slog         *DebugSlog
	inputHistory *DebugInputHistory
}

type DebugSlog struct {
	text       strings.Builder
	lineCount  int
	renderTime time.Duration
}

type DebugInputHistory struct {
	view      string
	itemCount int
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

func getInputHistory(m DebugCmd) string {
	histState := m.state.inputHistory
	if m.history.GetLen() == histState.itemCount {
		return histState.view
	}

	var sb strings.Builder
	items := m.history.GetInputs()
	for i, item := range items {
		if i == 0 {
			sb.WriteString(item)
			continue
		}
		sb.WriteString(fmt.Sprintf("\n%s", item))
	}
	histState.itemCount = m.history.GetLen()
	histState.view = histStyle.Render(sb.String())
	return histState.view
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
	return strings.TrimSpace(string(bytes))
}

func getSLog(m DebugCmd) string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	path := filepath.Join(dir, "log.txt")
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err.Error()
	}
	slog := m.state.slog

	str := strings.TrimSpace(string(bytes))
	lines := strings.Split(str, "\n")
	lineCount := len(lines)

	if lineCount == slog.lineCount {
		return fmt.Sprintf(
			"%s Took: %s",
			slog.text.String(),
			slog.renderTime,
		)
	}

	now := time.Now()
	defer func() {
		slog.renderTime = time.Since(now)
		slog.lineCount = lineCount
	}()

	if slog.lineCount > 0 {
		lines = lines[slog.lineCount:]
	}

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

		slog.text.WriteString(fmt.Sprintf("%s %s\n", tag, msg))
	}

	return slog.text.String()
}
