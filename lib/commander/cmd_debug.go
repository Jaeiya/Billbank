package commander

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
	"github.com/jaeiya/billbank/lib/ui"
	"github.com/jaeiya/billbank/lib/utils"
)

var (
	infoLogStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#29DEFF"))
	attnLogStyle  = lipgloss.NewStyle().Foreground(ui.FgWarnColor)
	errLogStyle   = lipgloss.NewStyle().Foreground(ui.FgErrLightColor)
	debugLogStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1BE697"))
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

func NewDebugCmd(h *utils.InputHistory) Command {
	m := debugCmdModel{
		BaseCommand: NewBaseCommand[debugCmdModel]([][]string{
			{"/", "x"},
			{"history", "log", "slog", "stats"},
			{"clear"},
		}),
		history: debugHistory{data: h},
		slog:    debugSlog{viewPort: viewport.New(0, 0)},
	}

	m.AddBranch([]BranchEntry[debugCmdModel]{
		{
			"/ history",
			func(dcm *debugCmdModel) tea.Cmd { return dcm.loadInputHistory() },
			func(dcm debugCmdModel) string { return dcm.viewHistory() },
		},
		{
			"/ log",
			func(dcm *debugCmdModel) tea.Cmd { return dcm.loadLog() },
			func(dcm debugCmdModel) string { return dcm.log.view },
		},
		{
			"/ slog",
			func(dcm *debugCmdModel) tea.Cmd { return dcm.loadSlog() },
			func(dcm debugCmdModel) string { return dcm.viewSlog() },
		},
		{
			"/ stats",
			func(dcm *debugCmdModel) tea.Cmd { return dcm.loadStats() },
			func(dcm debugCmdModel) string { return dcm.viewStats() },
		},
		{
			"/ log clear",
			func(dcm *debugCmdModel) tea.Cmd { return dcm.clearLog() },
			func(dcm debugCmdModel) string { return dcm.clearLogView() },
		},
		{
			"/ slog clear",
			func(dcm *debugCmdModel) tea.Cmd { return dcm.clearSlog() },
			func(dcm debugCmdModel) string { return dcm.clearSlogView() },
		},
	}...)

	return NewCommand(
		CommandConfig{
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
	*BaseCommand[debugCmdModel]
	history debugHistory
	log     debugLog
	slog    debugSlog
	stats   struct {
		data       debugStats
		memStats   runtime.MemStats
		renderTime time.Time
	}
}

func (m debugCmdModel) Update(msg tea.Msg) (CommandModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	model, cmd := m.BaseCommand.Update(&m, msg)
	m = *model
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+j" {
			m.slog.viewPort.LineDown(5)
		}
		if msg.String() == "ctrl+k" {
			m.slog.viewPort.LineUp(5)
		}
	}

	_, cmd = m.slog.viewPort.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m debugCmdModel) View() string {
	return m.BaseCommand.View(m)
}

func (m *debugCmdModel) loadInputHistory() tea.Cmd {
	if m.history.data.GetLen() == m.history.lastLen {
		return nil
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
	return nil
}

func (m debugCmdModel) viewHistory() string {
	return histStyle.Render(m.history.view)
}

func (m *debugCmdModel) loadLog() tea.Cmd {
	path := filepath.Join(utils.GetWorkingDir(), "log.txt")
	fileInfo, err := os.Stat(path)
	if err != nil {
		panic(err)
	}
	if fileInfo.Size() == int64(len(m.log.view)+1) {
		return nil
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		m.log.view = err.Error()
		return nil
	}
	m.log.view = strings.TrimSpace(string(bytes))
	m.log.lineCount = strings.Count(m.log.view, "\n")
	return nil
}

func (m *debugCmdModel) clearLog() tea.Cmd {
	path := filepath.Join(utils.GetWorkingDir(), "log.txt")
	err := os.Truncate(path, 0)
	if err != nil {
		panic(err)
	}
	m.log.view = ""
	m.log.lineCount = 0
	// No reason to hold old slog info
	m.slog.view = ""
	m.slog.lineCount = 0
	return nil
}

func (m debugCmdModel) clearLogView() string {
	return ui.NewInfoBox(
		"Clear Log",
		"The log has been successfully cleared!",
		m.viewWidth,
		m.viewHeight,
	)
}

func (m *debugCmdModel) loadSlog() tea.Cmd {
	m.loadLog()
	pollCmd := m.poll(time.Millisecond * 250)

	if m.log.lineCount == m.slog.lineCount {
		return pollCmd
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

	fixedWidthContent := lipgloss.NewStyle().Width(m.viewWidth - 1).Render(m.slog.view)
	m.slog.viewPort.SetContent(fixedWidthContent)
	m.slog.viewPort.GotoBottom()
	return pollCmd
}

func (m *debugCmdModel) clearSlog() tea.Cmd {
	m.slog.lineCount = 0
	m.slog.view = ""
	return nil
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

func (m *debugCmdModel) loadStats() tea.Cmd {
	var err error
	pollCmd := m.poll(time.Millisecond * 350)

	path := filepath.Join(utils.GetWorkingDir(), "log.txt")
	fileInfo, err := os.Stat(path)
	if err != nil {
		panic(err)
	}

	var historySize int
	for _, item := range m.history.data.GetInputs() {
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

	return pollCmd
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
