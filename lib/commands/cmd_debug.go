package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jaeiya/billbank/lib/utils"
)

var (
	infoLogStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#33B0FF"))
	attnLogStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFDE00"))
	errLogStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF82E9"))
	logStyle     = lipgloss.NewStyle().MarginLeft(1).MarginTop(1)
	msgStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#96F1D4"))
	histStyle    = lipgloss.NewStyle().Padding(1)
)

var cmdMap = map[string]func(DebugCmd) string{
	"history": getInputHistory,
	"log":     getLog,
	"slog":    getSLog,
}

func NewDebugCmd(h *utils.InputHistory) Command {
	vp := viewport.New(0, 0)
	dc := DebugCmd{
		inputHistory: h,
		state:        &DebugCmdState{},
		viewPort:     &vp,
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
				GetModel: func(status CommandStatus) utils.CommandModelMsg {
					dc.CommandStatus = status
					return dc
				},
				hasArg: false,
			},
		},
	)
}

type DebugCmdState struct {
	slogText       strings.Builder
	slogLineCount  int
	slogRenderTime time.Duration
	historyView    string
	historyCount   int
}

type DebugCmd struct {
	CommandStatus
	inputHistory   *utils.InputHistory
	state          *DebugCmdState
	viewPortWidth  int
	viewPortHeight int
	viewPort       *viewport.Model
}

func (DebugCmd) Init() tea.Cmd {
	return nil
}

func (cmd DebugCmd) Update(msg tea.Msg) (utils.CommandModelMsg, tea.Cmd) {
	var c tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+j" {
			cmd.viewPort.LineDown(5)
		}
		if msg.String() == "ctrl+k" {
			cmd.viewPort.LineUp(5)
		}

	case utils.ViewportSizeMsg:
		cmd.viewPortWidth = msg.Width
		cmd.viewPortHeight = msg.Height
		cmd.viewPort.Width = cmd.viewPortWidth
		cmd.viewPort.Height = cmd.viewPortHeight - 1
		cmd.viewPort.YPosition = 1
		utils.Log(utils.Info, msg.Height)
	}

	_, c = cmd.viewPort.Update(msg)
	cmds = append(cmds, c)

	return cmd, tea.Batch(cmds...)
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

func getInputHistory(cmd DebugCmd) string {
	if cmd.inputHistory.GetLen() == cmd.state.historyCount {
		return histStyle.Render(cmd.state.historyView)
	}

	var sb strings.Builder
	items := cmd.inputHistory.GetInputs()
	for i, item := range items {
		if i == 0 {
			sb.WriteString(item)
			continue
		}
		sb.WriteString(fmt.Sprintf("\n%s", item))
	}
	cmd.state.historyCount = cmd.inputHistory.GetLen()
	cmd.state.historyView = sb.String()

	return histStyle.Render(cmd.state.historyView)
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

func getSLog(cmd DebugCmd) string {
	str := getLog(cmd)
	lines := strings.Split(str, "\n")
	lineCount := len(lines)

	if lineCount == cmd.state.slogLineCount {
		content := fmt.Sprintf(
			"%s Took: %s",
			cmd.viewPort.View(),
			cmd.state.slogRenderTime,
		)
		return logStyle.Render(content)
	}

	now := time.Now()
	defer func() {
		cmd.state.slogRenderTime = time.Since(now)
		cmd.state.slogLineCount = lineCount
	}()

	if cmd.state.slogLineCount > 0 {
		lines = lines[cmd.state.slogLineCount:]
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

		switch tag {
		case "[NFO]":
			tag = infoLogStyle.Render(tag)
		case "[ATN]":
			tag = attnLogStyle.Render(tag)
			msg = attnLogStyle.Render(msg)
		case "[ERR]":
			tag = errLogStyle.Render(tag)
			msg = errLogStyle.Render(msg)
		}

		cmd.state.slogText.WriteString(fmt.Sprintf("%s %s\n", tag, msg))
	}

	cmd.viewPort.SetContent(cmd.state.slogText.String())
	cmd.viewPort.GotoBottom()
	return logStyle.Render(cmd.viewPort.View())
}
