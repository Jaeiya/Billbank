package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jaeiya/billbank/lib/cmd"
	"github.com/jaeiya/billbank/lib/ui"
	"github.com/jaeiya/billbank/lib/utils"
)

var (
	infoLogStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#29DEFF"))
	attnLogStyle  = lipgloss.NewStyle().Foreground(ui.FgWarnColor)
	errLogStyle   = lipgloss.NewStyle().Foreground(ui.FgErrLightColor)
	debugLogStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#2E3F00")).
			Foreground(lipgloss.Color("#D3FF5C"))
	hotLogStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#FF9600")).
			Foreground(lipgloss.Color("#000"))
	logStyle      = lipgloss.NewStyle().MarginLeft(1).MarginTop(1)
	slogWordStyle = lipgloss.NewStyle().Foreground(ui.FgColor)
	slogPathStyle = lipgloss.NewStyle().Align(lipgloss.Right).Foreground(ui.FgDimColor)
	histStyle     = lipgloss.NewStyle().Foreground(ui.FgColor).Padding(1)

	// Stats Styles
	statHeader = lipgloss.NewStyle().
			Width(30).
			Align(lipgloss.Center).
			Background(ui.BgDimColor).
			Foreground(ui.FgWarnColor).
			PaddingTop(1)

	statBox = lipgloss.NewStyle().
		Background(ui.BgDimColor).
		Foreground(ui.FgColor).
		Align(lipgloss.Left).
		Padding(1).
		PaddingLeft(2).
		Width(30)
)

func NewDebugCmd(h *utils.InputHistory) cmd.Command {
	vp := viewport.New(0, 0)
	vp.KeyMap.Down = key.NewBinding()
	vp.KeyMap.Up = key.NewBinding()
	vp.KeyMap.HalfPageUp = key.NewBinding(key.WithKeys("ctrl+k"))
	vp.KeyMap.HalfPageDown = key.NewBinding(key.WithKeys("ctrl+j"))

	m := debugModel{
		BaseCmdModel: cmd.NewBaseModel[debugModel]([][]string{
			{"/", "x"},
			{"history", "log", "slog", "stats"},
			{"clear"},
		}),
		history: debugHistory{data: h},
		slog:    debugSlog{viewPort: vp},
	}

	m.AddBranch([]cmd.Branch[debugModel]{
		{String: "/ history", Fn: loadHistory, ViewFn: viewHistory},
		{String: "/ log", Fn: loadLog, ViewFn: viewLog},
		{String: "/ slog", Fn: loadSlog, ViewFn: viewSlog},
		{String: "/ stats", Fn: loadStats, ViewFn: viewStats},
		{String: "/ log clear", Fn: clearLog, ViewFn: clearLogView},
		{String: "/ slog clear", Fn: clearSlog, ViewFn: clearSlogView},
	}...)

	return cmd.New(
		cmd.Config{
			Model:               m,
			InputValidationFunc: func(arg string) error { return nil },
			KeyValidationFunc:   func(key rune) bool { return false },
			HasArg:              false,
		},
	)
}

type debugModel struct {
	*cmd.BaseCmdModel[debugModel]
	history debugHistory
	log     debugLog
	slog    debugSlog
	stats   struct {
		data       debugStats
		memStats   runtime.MemStats
		renderTime time.Time
	}
}

type debugStats struct {
	historySize     uint64
	renderedLogSize uint64
	logSize         uint64
	slogSize        uint64
	memAlloc        uint64
	memTotal        uint64
	memWorking      uint64
	memGcCount      uint64
}

type debugHistory struct {
	view    string
	lastLen int
	data    *utils.InputHistory
}

type debugLog struct {
	view      string
	lineCount int
}

type debugSlog struct {
	debugLog
	viewPort      viewport.Model
	lastRenderDur time.Duration
}

func (m debugModel) Update(msg tea.Msg) (cmd.Model, tea.Cmd) {
	var teaCmd tea.Cmd
	var teaCmds []tea.Cmd

	m, teaCmd = m.BaseCmdModel.Update(m, msg)
	teaCmds = append(teaCmds, teaCmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.IsActiveBranch("/ slog") {
			// Reload slog
			if msg.String() == "ctrl+r" {
				m = m.Exec(m)
			}
		}
	}

	m.slog.viewPort, teaCmd = m.slog.viewPort.Update(msg)
	teaCmds = append(teaCmds, teaCmd)

	return m, tea.Batch(teaCmds...)
}

func (m debugModel) View() string {
	return m.BaseCmdModel.View(m)
}

func loadHistory(m debugModel) debugModel {
	if m.history.data.GetLen() == m.history.lastLen {
		return m
	}

	var sb strings.Builder
	list := m.history.data.ListHistory()
	for i, item := range list {
		if i == 0 {
			sb.WriteString(item)
			continue
		}
		sb.WriteString(fmt.Sprintf("\n%s", item))
	}
	m.history.lastLen = m.history.data.GetLen()
	m.history.view = sb.String()
	return m
}

func viewHistory(m debugModel) string {
	return histStyle.Render(m.history.view)
}

func loadLog(m debugModel) debugModel {
	path := filepath.Join(utils.GetWorkingDir(), "log.txt")
	fileInfo, err := os.Stat(path)
	if err != nil {
		m.AddError(err)
		return m
	}
	if fileInfo.Size() == int64(len(m.log.view)+1) {
		return m
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		m.AddError(err)
		return m
	}
	m.log.view = strings.TrimSpace(string(bytes))
	m.log.lineCount = strings.Count(m.log.view, "\n")
	return m
}

func viewLog(m debugModel) string {
	return m.log.view
}

func clearLog(m debugModel) debugModel {
	path := filepath.Join(utils.GetWorkingDir(), "log.txt")
	err := os.Truncate(path, 0)
	if err != nil {
		m.AddError(err)
		return m
	}
	m.log.view = ""
	m.log.lineCount = 0
	// No reason to hold old slog info
	m.slog.view = ""
	m.slog.lineCount = 0
	return m
}

func clearLogView(m debugModel) string {
	w, h := m.GetViewSize()
	return ui.NewInfoBox(
		"Clear Log",
		"The log has been successfully cleared!",
		w, h,
	)
}

func loadSlog(m debugModel) debugModel {
	m = loadLog(m)

	if m.log.lineCount == m.slog.lineCount {
		return m
	}

	var tagBuilder, pathBuilder, wordBuilder strings.Builder
	now := time.Now()

	lines := strings.Split(m.log.view, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
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
		case "[HOT]":
			tag = hotLogStyle.Render(tag)
			subjectStyle = hotLogStyle
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

	content := strings.TrimSpace(lipgloss.JoinHorizontal(
		lipgloss.Left,
		tagBuilder.String(),
		slogPathStyle.Render(pathBuilder.String()),
		wordBuilder.String(),
	))

	m.slog.lastRenderDur = time.Since(now)
	m.slog.lineCount = m.log.lineCount
	m.slog.view = content

	// The terminal can be resized at any time
	w, h := m.GetViewSize()
	m.slog.viewPort.Width = w
	m.slog.viewPort.Height = h - 2

	fixedWidthContent := lipgloss.NewStyle().Width(w - 1).Render(m.slog.view)
	m.slog.viewPort.SetContent(fixedWidthContent)
	m.slog.viewPort.GotoBottom()
	return m
}

func clearSlog(m debugModel) debugModel {
	m.slog.lineCount = 0
	m.slog.view = ""
	return m
}

func clearSlogView(m debugModel) string {
	w, h := m.GetViewSize()
	return ui.NewInfoBox(
		"Clear Slog",
		"Slog has been reset and will be re-rendered on execution.",
		w, h,
	)
}

func viewSlog(m debugModel) string {
	content := fmt.Sprintf(
		"%s\nTook: %s",
		m.slog.viewPort.View(),
		m.slog.lastRenderDur,
	)
	return logStyle.Render(content)
}

func loadStats(m debugModel) debugModel {
	var err error

	path := filepath.Join(utils.GetWorkingDir(), "log.txt")
	fileInfo, err := os.Stat(path)
	if err != nil {
		m.AddError(err)
		return m
	}

	var historySize int
	for _, item := range m.history.data.ListHistory() {
		historySize += len(item)
	}

	runtime.ReadMemStats(&m.stats.memStats)

	m.stats.data = debugStats{
		historySize:     uint64(historySize),
		renderedLogSize: uint64(len(m.log.view)),
		logSize:         uint64(fileInfo.Size()),
		slogSize:        uint64(len(m.slog.view)),
		memGcCount:      uint64(m.stats.data.memAlloc),
		memAlloc:        m.stats.memStats.Alloc,
		memTotal:        m.stats.memStats.Sys,
		// Simulate what task manager provides as the working memory
		memWorking: m.stats.memStats.OtherSys + m.stats.memStats.HeapSys,
	}

	return m
}

func viewStats(m debugModel) string {
	debugHeader := statHeader.Render("History Stats")

	debugTitles := lipgloss.JoinVertical(
		lipgloss.Right,
		"Items: ",
		"Size: ",
	)

	debugValues := lipgloss.JoinVertical(
		lipgloss.Left,
		strconv.Itoa(m.history.data.GetLen()),
		utils.FormatBytes(m.stats.data.historySize),
	)

	debugStats := statBox.PaddingBottom(0).Render(lipgloss.JoinHorizontal(
		lipgloss.Left,
		debugTitles,
		debugValues,
	))

	logHeader := statHeader.Render("Log Stats")

	logTitles := lipgloss.JoinVertical(
		lipgloss.Right,
		"Size: ",
		"Slog: ",
		"Lines: ",
		"Cache: ",
	)

	logValues := lipgloss.JoinVertical(
		lipgloss.Left,
		utils.FormatBytes(m.stats.data.logSize),
		utils.FormatBytes(m.stats.data.slogSize),
		strconv.Itoa(m.log.lineCount),
		utils.FormatBytes(m.stats.data.renderedLogSize+m.stats.data.slogSize),
	)

	logStats := statBox.Render(lipgloss.JoinHorizontal(
		lipgloss.Left,
		logTitles,
		logValues,
	))

	memHeader := statHeader.Render("Memory Stats")

	memTitles := lipgloss.JoinVertical(
		lipgloss.Right,
		"Allocated: ",
		"Working: ",
		"Total: ",
		"GC Count: ",
	)

	memValues := lipgloss.JoinVertical(
		lipgloss.Left,
		utils.FormatBytes(m.stats.data.memAlloc),
		utils.FormatBytes(m.stats.data.memWorking),
		utils.FormatBytes(m.stats.data.memTotal),
		strconv.FormatUint(m.stats.data.memGcCount, 10),
	)

	memStats := statBox.Render(lipgloss.JoinHorizontal(
		lipgloss.Left,
		memTitles,
		memValues,
	))

	w, h := m.GetViewSize()

	return lipgloss.Place(
		w, h,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			lipgloss.JoinVertical(lipgloss.Left, debugHeader, debugStats, logHeader, logStats),
			" ",
			lipgloss.JoinVertical(lipgloss.Left, memHeader, memStats),
		),
	)
}
