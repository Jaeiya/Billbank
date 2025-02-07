package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jaeiya/billbank/lib"
	"github.com/jaeiya/billbank/lib/ui"
	"github.com/jaeiya/billbank/lib/utils"
)

var (
	infoLogStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#29DEFF"))
	attnLogStyle  = lipgloss.NewStyle().Foreground(lib.FgWarnColor)
	errLogStyle   = lipgloss.NewStyle().Foreground(lib.FgErrLightColor)
	debugLogStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1BE697"))
	logStyle      = lipgloss.NewStyle().MarginLeft(1).MarginTop(1)
	slogWordStyle = lipgloss.NewStyle().Foreground(lib.FgColor)
	slogPathStyle = lipgloss.NewStyle().Align(lipgloss.Right).Foreground(lib.FgDimColor)
	histStyle     = lipgloss.NewStyle().Foreground(lib.FgColor).Padding(1)

	// Stats Styles
	statHeader = lipgloss.NewStyle().
			Width(30).
			Align(lipgloss.Center).
			Background(lib.BgDimColor).
			Foreground(lib.FgWarnColor).
			PaddingTop(1)

	statBox = lipgloss.NewStyle().
		Background(lib.BgDimColor).
		Foreground(lib.FgColor).
		Align(lipgloss.Left).
		Padding(1).
		PaddingLeft(2).
		Width(30)
)

func NewDebugCmd(h *utils.InputHistory) ui.Command {
	m := debugCmdModel{
		BaseCommand: NewBaseCommand[debugCmdModel]([][]string{
			{"/"},
			{"history", "log", "slog", "stats"},
			{"clear"},
		}),
		history: debugHistory{data: h},
		slog:    debugSlog{viewPort: viewport.New(0, 0)},
	}

	m.AddCommands([]CommandEntry[debugCmdModel]{
		{
			"/ history",
			func(dcm debugCmdModel) debugCmdModel { return dcm.loadInputHistory() },
			func(dcm debugCmdModel) string { return dcm.viewHistory() },
		},
		{
			"/ log",
			func(dcm debugCmdModel) debugCmdModel { return dcm.loadLog() },
			func(dcm debugCmdModel) string { return dcm.log.view },
		},
		{
			"/ slog",
			func(dcm debugCmdModel) debugCmdModel { return dcm.loadSlog() },
			func(dcm debugCmdModel) string { return dcm.viewSlog() },
		},
		{
			"/ stats",
			func(dcm debugCmdModel) debugCmdModel { return dcm.loadStats() },
			func(dcm debugCmdModel) string { return dcm.viewStats() },
		},
		{
			"/ log clear",
			func(dcm debugCmdModel) debugCmdModel { return dcm.clearLog() },
			func(dcm debugCmdModel) string { return dcm.clearLogView() },
		},
		{
			"/ slog clear",
			func(dcm debugCmdModel) debugCmdModel { return dcm.clearSlog() },
			func(dcm debugCmdModel) string { return dcm.clearSlogView() },
		},
	}...)

	return ui.NewCommand(
		ui.CommandConfig{
			Model:               m,
			InputValidationFunc: func(arg string) error { return nil },
			KeyValidationFunc:   func(key rune) bool { return false },
			HasArg:              false,
		},
	)
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

type debugCmdModel struct {
	BaseCommand[debugCmdModel]
	history debugHistory
	log     debugLog
	slog    debugSlog
	stats   struct {
		data       debugStats
		renderTime time.Time
	}
}

func (m debugCmdModel) Update(msg tea.Msg) (ui.CommandModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	m.BaseCommand.Update(msg)

	if m.GetError() != nil {
		m.SetError(nil)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+j" {
			m.slog.viewPort.LineDown(5)
		}
		if msg.String() == "ctrl+k" {
			m.slog.viewPort.LineUp(5)
		}
	}

	m, err := m.Exec(m)
	if err != nil {
		m.SetError(err)
	} else if !m.HasView() {
		m.SetError(fmt.Errorf("tried to display missing view from [%s]", m.cmdStatus.TreeStr))
	}

	_, cmd = m.slog.viewPort.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m debugCmdModel) View() string {
	return m.ExecView(m)
}

func (m debugCmdModel) SetStatus(status ui.CommandStatus) ui.CommandModel {
	m.BaseCommand.SetStatus(status)
	return m
}

func (m debugCmdModel) loadInputHistory() debugCmdModel {
	if m.history.data.GetLen() == m.history.lastLen {
		return m
	}

	var sb strings.Builder
	items := m.history.data.GetInputs()
	for i, item := range items {
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

func (m debugCmdModel) viewHistory() string {
	return histStyle.Render(m.history.view)
}

func (m debugCmdModel) loadLog() debugCmdModel {
	path := filepath.Join(utils.GetWorkingDir(), "log.txt")
	fileInfo, err := os.Stat(path)
	if err != nil {
		panic(err)
	}
	if fileInfo.Size() == int64(len(m.log.view)+1) {
		return m
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		m.log.view = err.Error()
		return m
	}
	m.log.view = strings.TrimSpace(string(bytes))
	m.log.lineCount = strings.Count(m.log.view, "\n")
	return m
}

func (m debugCmdModel) clearLog() debugCmdModel {
	path := filepath.Join(utils.GetWorkingDir(), "log.txt")
	err := os.Truncate(path, 0)
	if err != nil {
		panic(err)
	}
	m.log.view = ""
	m.slog.view = ""
	m.slog.lineCount = 0
	return m
}

func (m debugCmdModel) clearLogView() string {
	return ui.NewInfoBox(
		"Clear Log",
		"The log has been successfully cleared!",
		m.viewWidth,
		m.viewHeight,
	)
}

func (m debugCmdModel) loadSlog() debugCmdModel {
	m = m.loadLog()
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
	m.slog.viewPort.Width = m.viewWidth
	m.slog.viewPort.Height = m.viewHeight - 2

	m.slog.viewPort.SetContent(lipgloss.NewStyle().Width(m.viewWidth - 1).Render(m.slog.view))
	m.slog.viewPort.GotoBottom()
	return m
}

func (m debugCmdModel) clearSlog() debugCmdModel {
	m.slog.lineCount = 0
	m.slog.view = ""
	return m
}

func (m debugCmdModel) clearSlogView() string {
	return ui.NewInfoBox(
		"Clear Slog",
		"Slog has been reset and will be re-rendered on execution.",
		m.viewWidth,
		m.viewHeight,
	)
}

func (m debugCmdModel) viewSlog() string {
	content := fmt.Sprintf(
		"%s\nTook: %s",
		m.slog.viewPort.View(),
		m.slog.lastRenderDur,
	)
	return logStyle.Render(content)
}

func (m debugCmdModel) loadStats() debugCmdModel {
	var err error
	now := time.Now()
	ms := now.Sub(m.stats.renderTime).Milliseconds()

	if ms < 1000 {
		return m
	}

	m.stats.renderTime = now

	path := filepath.Join(utils.GetWorkingDir(), "log.txt")
	fileInfo, err := os.Stat(path)
	if err != nil {
		panic(err)
	}

	var historySize int
	for _, item := range m.history.data.GetInputs() {
		historySize += len(item)
	}

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	m.stats.data = debugStats{
		historySize:     uint64(historySize),
		renderedLogSize: uint64(len(m.log.view)),
		logSize:         uint64(fileInfo.Size()),
		slogSize:        uint64(len(m.slog.view)),
		memGcCount:      uint64(mem.NumGC),
		memAlloc:        mem.Alloc,
		memTotal:        mem.Sys,
		// Simulate what task manager provides as the working memory
		memWorking: mem.OtherSys + mem.HeapSys,
	}

	return m
}

func (m debugCmdModel) viewStats() string {
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

	return lipgloss.Place(
		m.viewWidth,
		m.viewHeight,
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
