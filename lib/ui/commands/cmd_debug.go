package ui

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
		slogViewPort: vp,
	}

	return NewCommand(
		CommandConfig{
			Command: Command{
				tree: dc.GetCmdTree(),
				GetModel: func(status CommandStatus) CommandModel {
					dc.CommandStatus = status
					return dc
				},
				hasArg: false,
			},
		},
	)
}

type DebugCmdModel struct {
	CommandStatus
	inputHistory   *utils.InputHistory
	historyLen     int
	historyView    string
	lastLogStr     string
	lastSlogStr    string
	slogViewPort   viewport.Model
	slogLineCount  int
	slogRenderTime time.Duration
}

func (DebugCmdModel) Init() tea.Cmd {
	return nil
}

func (m DebugCmdModel) Update(msg tea.Msg) (CommandModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+j" {
			m.slogViewPort.LineDown(5)
		}
		if msg.String() == "ctrl+k" {
			m.slogViewPort.LineUp(5)
		}

	case utils.ViewportSizeMsg:
		m.slogViewPort.Width = msg.Width
		m.slogViewPort.Height = msg.Height - 2

	case utils.CommandStrMsg:
		m.CommandStr = string(msg)
	}

	switch m.CommandStr {
	case "log":
		m.loadLog()

	case "history":
		m.loadInputHistory()

	case "slog":
		m.loadSlog()

	}

	_, cmd = m.slogViewPort.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m DebugCmdModel) View() string {
	switch m.CommandStr {
	case "history":
		return histStyle.Render(m.historyView)

	case "log":
		return m.lastLogStr

	case "slog":
		content := fmt.Sprintf(
			"%s\nTook: %s",
			m.slogViewPort.View(),
			m.slogRenderTime,
		)
		return logStyle.Render(content)

	default:
		return "command has no view"
	}
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

func (m *DebugCmdModel) loadInputHistory() {
	if m.inputHistory.GetLen() == m.historyLen {
		return
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
	m.historyLen = m.inputHistory.GetLen()
	m.historyView = sb.String()
}

func (m *DebugCmdModel) loadLog() {
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
	if fileInfo.Size() == int64(len(m.lastLogStr)+1) {
		return
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		m.lastLogStr = err.Error()
		return
	}
	m.lastLogStr = strings.TrimSpace(string(bytes))
}

func (m *DebugCmdModel) loadSlog() {
	m.loadLog()
	lines := strings.Split(m.lastLogStr, "\n")
	lineCount := len(lines)

	if lineCount == m.slogLineCount {
		return
	}

	now := time.Now()
	defer func() {
		m.slogRenderTime = time.Since(now)
		m.slogLineCount = lineCount
	}()

	var sb strings.Builder
	if m.slogLineCount > 0 {
		sb.WriteString(strings.TrimSpace(m.lastSlogStr) + "\n")
		lines = lines[m.slogLineCount:]
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

		sb.WriteString(fmt.Sprintf("%s %s\n", tag, msg))
	}

	m.lastSlogStr = sb.String()
	m.slogViewPort.SetContent(sb.String())
	m.slogViewPort.GotoBottom()
}
