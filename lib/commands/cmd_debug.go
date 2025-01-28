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

func NewDebugCmd(h *utils.InputHistory) Command {
	vp := viewport.New(0, 0)
	vp.YPosition = 1
	dc := DebugCmdModel{
		inputHistory: h,
		state:        &DebugCmdState{},
		viewPort:     vp,
	}

	return NewCommand(
		CommandConfig{
			Command: Command{
				tree: dc.GetCmdTree(),
				GetModel: func(status CommandStatus) utils.CommandModel {
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
	lastLogStr     string
}

type DebugCmdModel struct {
	CommandStatus
	inputHistory   *utils.InputHistory
	state          *DebugCmdState
	viewPortWidth  int
	viewPortHeight int
	viewPort       viewport.Model
}

func (DebugCmdModel) Init() tea.Cmd {
	return nil
}

func (m DebugCmdModel) Update(msg tea.Msg) (utils.CommandModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+j" {
			m.viewPort.LineDown(5)
		}
		if msg.String() == "ctrl+k" {
			m.viewPort.LineUp(5)
		}

	case utils.ViewportSizeMsg:
		m.viewPortWidth = msg.Width
		m.viewPortHeight = msg.Height
		m.viewPort.Width = m.viewPortWidth
		m.viewPort.Height = m.viewPortHeight - 2
	}

	switch m.CommandStr {
	case "log":
		m, _ = m.cacheLog()

	case "history":
		m, _ = m.getInputHistory()

	case "slog":
		m, _ = m.cacheSlog()

	}

	_, cmd = m.viewPort.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m DebugCmdModel) View() string {
	switch m.CommandStr {
	case "history":
		return histStyle.Render(m.state.historyView)
	case "log":
		return m.state.lastLogStr
	case "slog":
		content := fmt.Sprintf(
			"%s\nTook: %s",
			m.viewPort.View(),
			m.state.slogRenderTime,
		)
		return logStyle.Render(content)
	default:
		return "command has no view"
	}

	// panic("missing view; use IsImplemented to check for this error")
}

func (m DebugCmdModel) IsImplemented() bool {
	return true
}

func (DebugCmdModel) GetCmdTree() [][]string {
	return [][]string{
		{"/"},
		{"history", "log", "slog"},
	}
}

func (m DebugCmdModel) getInputHistory() (DebugCmdModel, tea.Cmd) {
	if m.inputHistory.GetLen() == m.state.historyCount {
		return m, nil
	}

	var sb strings.Builder
	items := m.inputHistory.GetInputs()
	for i, item := range items {
		if i == 0 {
			sb.WriteString(item)
			continue
		}
		sb.WriteString(fmt.Sprintf("\n%s", item))
	}
	m.state.historyCount = m.inputHistory.GetLen()
	m.state.historyView = sb.String()

	return m, nil
}

func (m DebugCmdModel) cacheLog() (DebugCmdModel, tea.Cmd) {
	// TODO - refactor this into utils so we can cache result
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	path := filepath.Join(dir, "log.txt")
	fileInfo, err := os.Stat(path)
	if err != nil {
		panic(err)
	}
	state := m.state
	if fileInfo.Size() == int64(len(state.lastLogStr)+1) {
		return m, nil
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		state.lastLogStr = err.Error()
		return m, nil
	}
	state.lastLogStr = strings.TrimSpace(string(bytes))
	return m, nil
}

func (m DebugCmdModel) cacheSlog() (DebugCmdModel, tea.Cmd) {
	m.cacheLog()
	state := m.state
	lines := strings.Split(state.lastLogStr, "\n")
	lineCount := len(lines)

	if lineCount == m.state.slogLineCount {
		return m, nil
	}

	now := time.Now()
	defer func() {
		m.state.slogRenderTime = time.Since(now)
		m.state.slogLineCount = lineCount
	}()

	if m.state.slogLineCount > 0 {
		lines = lines[m.state.slogLineCount:]
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

		m.state.slogText.WriteString(fmt.Sprintf("%s %s\n", tag, msg))
	}

	m.viewPort.SetContent(m.state.slogText.String())
	m.viewPort.GotoBottom()
	return m, nil
}
