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
	"github.com/jaeiya/billbank/lib"
	"github.com/jaeiya/billbank/lib/ui"
	"github.com/jaeiya/billbank/lib/utils"
)

type debugCmdTree string

const (
	DebugHistory = debugCmdTree("/ history")
	DebugLog     = debugCmdTree("/ log")
	DebugSlog    = debugCmdTree("/ slog")
)

var (
	infoLogStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#29DEFF"))
	attnLogStyle  = lipgloss.NewStyle().Foreground(lib.FgWarnColor)
	errLogStyle   = lipgloss.NewStyle().Foreground(lib.FgErrLightColor)
	debugLogStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#D283FF"))
	logStyle      = lipgloss.NewStyle().MarginLeft(1).MarginTop(1)
	slogWordStyle = lipgloss.NewStyle().Foreground(lib.FgColor)
	slogPathStyle = lipgloss.NewStyle().Align(lipgloss.Right).Foreground(lib.FgDimColor)
	histStyle     = lipgloss.NewStyle().Foreground(lib.FgColor).Padding(1)
)

type (
	debugCmdMap     map[debugCmdTree]func(debugCmdModel) debugCmdModel
	debugCmdViewMap map[debugCmdTree]func(debugCmdModel) string
)

func NewDebugCmd(h *utils.InputHistory) ui.Command {
	vp := viewport.New(0, 0)
	vp.YPosition = 1

	m := debugCmdModel{
		inputHistory: h,
		slogViewPort: vp,
	}

	m.cmdMap = debugCmdMap{
		DebugHistory: func(dcm debugCmdModel) debugCmdModel { return dcm.loadInputHistory() },
		DebugLog:     func(dcm debugCmdModel) debugCmdModel { return dcm.loadLog() },
		DebugSlog:    func(dcm debugCmdModel) debugCmdModel { return dcm.loadSlog() },
	}

	m.cmdViewMap = debugCmdViewMap{
		DebugHistory: func(dcm debugCmdModel) string { return dcm.viewHistory() },
		DebugLog:     func(dcm debugCmdModel) string { return dcm.lastLogStr },
		DebugSlog:    func(dcm debugCmdModel) string { return dcm.viewSlog() },
	}

	return ui.NewCommand(
		ui.CommandConfig{
			Tree:                m.GetCmdTree(),
			Model:               m,
			InputValidationFunc: func(arg string) error { return nil },
			KeyValidationFunc:   func(key rune) bool { return false },
			HasArg:              false,
		},
	)
}

type debugCmdModel struct {
	viewportWidth  int
	viewportHeight int
	cmdStatus      ui.CommandStatus
	cmdMap         debugCmdMap
	cmdViewMap     debugCmdViewMap
	lastError      error
	inputHistory   *utils.InputHistory
	historyLen     int
	historyView    string
	lastLogStr     string
	lastSlogStr    string
	slogViewPort   viewport.Model
	slogLineCount  int
	slogRenderTime time.Duration
}

func (m debugCmdModel) Update(msg tea.Msg) (ui.CommandModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	if m.lastError != nil {
		m.lastError = nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+j" {
			m.slogViewPort.LineDown(5)
		}
		if msg.String() == "ctrl+k" {
			m.slogViewPort.LineUp(5)
		}

	case ui.ViewportSizeMsg:
		m.viewportWidth = msg.Width
		m.viewportHeight = msg.Height
		m.slogViewPort.Width = msg.Width

		m.slogViewPort.Height = msg.Height - 2
	}

	exec, hasCmd := m.cmdMap[debugCmdTree(m.cmdStatus.TreeStr)]
	if hasCmd {
		m = exec(m)
	} else {
		m.lastError = fmt.Errorf("'%s' not implemented", m.cmdStatus.TreeStr)
	}

	_, cmd = m.slogViewPort.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m debugCmdModel) View() string {
	viewFunc, hasView := m.cmdViewMap[debugCmdTree(m.cmdStatus.TreeStr)]
	if hasView {
		return viewFunc(m)
	}
	return fmt.Sprintf("Missing View for Command::%s", m.cmdStatus.TreeStr)
}

func (m debugCmdModel) SetStatus(status ui.CommandStatus) ui.CommandModel {
	m.cmdStatus = status
	return m
}

func (m debugCmdModel) IsTreeSupported(treeStr string) bool {
	switch debugCmdTree(treeStr) {
	case DebugHistory, DebugLog, DebugSlog:
		return true
	}
	return false
}

func (debugCmdModel) GetCmdTree() [][]string {
	return [][]string{
		{"/"},
		{"history", "log", "slog"},
	}
}

func (m debugCmdModel) GetLastError() error {
	return m.lastError
}

func (m debugCmdModel) loadInputHistory() debugCmdModel {
	if m.inputHistory.GetLen() == m.historyLen {
		return m
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
	return m
}

func (m debugCmdModel) viewHistory() string {
	return histStyle.Render(m.historyView)
}

func (m debugCmdModel) loadLog() debugCmdModel {
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
		return m
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		m.lastLogStr = err.Error()
		return m
	}
	m.lastLogStr = strings.TrimSpace(string(bytes))
	return m
}

func (m debugCmdModel) loadSlog() debugCmdModel {
	m = m.loadLog()
	lines := strings.Split(m.lastLogStr, "\n")
	lineCount := len(lines)

	if lineCount == m.slogLineCount {
		return m
	}

	now := time.Now()

	var lastLog string
	if m.slogLineCount > 0 {
		lastLog = strings.TrimSpace(m.lastSlogStr) + "\n"
		lines = lines[m.slogLineCount:]
	}

	var tagBuilder, pathBuilder, wordBuilder strings.Builder
	for _, line := range lines {
		var parts []string = strings.Split(line, " ")
		var tag string = parts[3]
		var subjectStyle lipgloss.Style

		switch tag {
		case "[NFO]":
			tag = infoLogStyle.Render(tag)
			subjectStyle = infoLogStyle
		case "[ATN]":
			tag = attnLogStyle.Render(tag)
			subjectStyle = attnLogStyle
		case "[ERR]":
			tag = errLogStyle.Render(tag)
			subjectStyle = errLogStyle
		case "[DBG]":
			tag = debugLogStyle.Render(tag)
			subjectStyle = debugLogStyle
		}

		tagBuilder.WriteString(fmt.Sprintf("%s \n", tag))

		var msg string
		var words []string = parts[5:]
		if strings.Contains(words[0], ":") {
			words[0] = subjectStyle.Render(words[0])
			words[1] = slogWordStyle.Render(strings.Join(words[1:], " "))
			msg = fmt.Sprintf("%s %s", words[0], words[1])
		} else {
			msg = slogWordStyle.Render(strings.Join(words, " "))
		}

		pathBuilder.WriteString(fmt.Sprintf("%s] \n", strings.Split(parts[4], ".")[0]))
		wordBuilder.WriteString(fmt.Sprintf("%s\n", msg))
	}

	lastLog = strings.TrimSpace(lastLog)
	content := strings.TrimSpace(lipgloss.JoinHorizontal(
		lipgloss.Left,
		tagBuilder.String(),
		slogPathStyle.Render(pathBuilder.String()),
		wordBuilder.String(),
	))

	if len(lastLog) > 0 {
		// We need to remove the special formatting that
		// lipgloss applies when aligning.
		formattingOffset := 46
		content = lipgloss.JoinVertical(
			lipgloss.Top,
			lastLog[:len(lastLog)-formattingOffset],
			content,
		)
	}
	m.slogRenderTime = time.Since(now)
	m.slogLineCount = lineCount
	m.lastSlogStr = content
	m.slogViewPort.SetContent(m.lastSlogStr)
	m.slogViewPort.GotoBottom()
	return m
}

func (m debugCmdModel) viewSlog() string {
	content := fmt.Sprintf(
		"%s\nTook: %s",
		m.slogViewPort.View(),
		m.slogRenderTime,
	)
	return logStyle.Render(content)
}
