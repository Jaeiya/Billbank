package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/jaeiya/billbank/internal/cmdcore"
	"github.com/jaeiya/billbank/internal/logger"
	"github.com/jaeiya/billbank/internal/ui"
	"github.com/jaeiya/billbank/internal/utils"
)

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
	data    *cmdcore.InputHistory
	view    string
	lastLen int
}

type debugLog struct {
	view      string
	lineCount int
}

type debugSlog struct {
	viewPort      viewport.Model
	lastRenderDur time.Duration
	view          string
}

var (
	infoLogStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#29DEFF"))
	attnLogStyle = lipgloss.NewStyle().Background(ui.BgDarkColor).Foreground(ui.FgWarnColor)
	errLogStyle  = lipgloss.NewStyle().
			Background(ui.BgDarkColor).
			Foreground(ui.FgErrColor)
	debugLogStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#2E3F00")).
			Foreground(lipgloss.Color("#D3FF5C"))
	hotLogStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#FF9600")).
			Foreground(lipgloss.Color("#000"))
	insaneLogStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#FF004E")).
			Foreground(lipgloss.Color("#FFCFDE"))
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

type debugModel struct {
	cmdcore.ModelBase
	history debugHistory
	log     debugLog
	slog    debugSlog
	stats   struct {
		data     debugStats
		memStats runtime.MemStats
	}
}

func NewDebugHandler(h *cmdcore.InputHistory) cmdcore.CommandHandler {
	vp := viewport.New(viewport.WithHeight(0), viewport.WithWidth(0))
	vp.KeyMap.Down = key.NewBinding(key.WithKeys("j"))
	vp.KeyMap.Up = key.NewBinding(key.WithKeys("k"))
	vp.KeyMap.PageDown = key.NewBinding()
	vp.KeyMap.PageUp = key.NewBinding()
	vp.KeyMap.HalfPageUp = key.NewBinding(key.WithKeys("ctrl+k"))
	vp.KeyMap.HalfPageDown = key.NewBinding(key.WithKeys("ctrl+j"))

	m := debugModel{
		ModelBase: cmdcore.NewModelBase(),
		history:   debugHistory{data: h},
		slog:      debugSlog{viewPort: vp},
	}

	commands := []cmdcore.Command[debugModel]{
		cmdcore.NewCommand(cmdcore.CommandOptions[debugModel, cmdcore.NoArg]{
			Path:     "history",
			RunFunc:  loadHistory,
			ViewFunc: viewHistory,
		}),

		cmdcore.NewCommand(cmdcore.CommandOptions[debugModel, int]{
			Path:      "slog",
			ArgType:   cmdcore.ArgOptional,
			RunFunc:   loadSlog,
			ViewFunc:  viewSlog,
			ParseFunc: validateSlogInput,
		}),

		cmdcore.NewCommand(cmdcore.CommandOptions[debugModel, cmdcore.NoArg]{
			Path:     "stats",
			RunFunc:  loadStats,
			ViewFunc: viewStats,
		}),

		cmdcore.NewCommand(cmdcore.CommandOptions[debugModel, cmdcore.NoArg]{
			Path:     "clear slog",
			RunFunc:  clearLog,
			ViewFunc: clearSlogView,
		}),

		cmdcore.NewCommand(cmdcore.CommandOptions[debugModel, logger.LogLevel]{
			Path:     "log_level",
			ArgType:  cmdcore.ArgRequired,
			RunFunc:  setLogLevel,
			ViewFunc: viewLogLevel,
			ParseFunc: func(arg string) (logger.LogLevel, error) {
				v, err := utils.ParseInt(arg)
				if err != nil {
					return 0, err
				}

				ll := logger.LogLevel(v)
				if !ll.IsValid() {
					return 0, fmt.Errorf("'%s' is not a valid log level", arg)
				}

				return ll, nil
			},
		}),
	}

	return cmdcore.NewCmdHandler(
		"debug",
		[]string{"/"},
		commands,
		m,
	)
}

func (m debugModel) Update(msg tea.Msg) (cmdcore.CommandModel, tea.Cmd) {
	var teaCmd tea.Cmd
	var teaCmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		// Reload slog
		if msg.String() == "ctrl+r" {
			// Wait for log to be written
			time.Sleep(10 * time.Millisecond)
			teaCmds = append(teaCmds, m.Reload("/ slog"))
		}
	}

	m.slog.viewPort, teaCmd = m.slog.viewPort.Update(msg)
	teaCmds = append(teaCmds, teaCmd)
	return m, tea.Batch(teaCmds...)
}

func loadHistory(m debugModel, arg *cmdcore.NoArg) (cmdcore.CommandModel, error) {
	if m.history.data.GetLen() == m.history.lastLen {
		return m, nil
	}

	var sb strings.Builder
	list := m.history.data.ListHistory()
	for i, item := range list {
		if i == 0 {
			sb.WriteString(item)
			continue
		}
		fmt.Fprintf(&sb, "\n%s", item)
	}
	m.history.lastLen = m.history.data.GetLen()
	m.history.view = sb.String()
	return m, nil
}

func viewHistory(m debugModel) tea.View {
	return tea.NewView(histStyle.Render(m.history.view))
}

func clearLog(m debugModel, arg *cmdcore.NoArg) (cmdcore.CommandModel, error) {
	err := logger.Reset()
	if err != nil {
		return m, err
	}
	m.log.view = ""
	m.log.lineCount = 0
	m.slog.view = ""
	return m, nil
}

func validateSlogInput(arg string) (int, error) {
	v, err := utils.ParseInt(arg)
	if err != nil {
		return 0, fmt.Errorf("[%s] is not a valid number of lines", arg)
	}
	return v, nil
}

func loadSlog(m debugModel, arg *int) (cmdcore.CommandModel, error) {
	path := filepath.Join(utils.GetWorkingDir(), "log.txt")
	fileInfo, err := os.Stat(path)
	if err != nil {
		return m, err
	}

	// Check if log is already cached
	if fileInfo.Size() == int64(len(m.log.view)+1) {
		logger.Log(logger.Debug, "[loadSlog] log is already cached")
		return m, nil
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return m, err
	}

	m.log.view = strings.TrimSpace(string(bytes))
	m.log.lineCount = strings.Count(m.log.view, "\n")

	maxLines := 0
	if arg != nil {
		maxLines = *arg
	}

	var tagBuilder, subjBuilder, wordBuilder, timeBuilder strings.Builder
	now := time.Now()

	lines := strings.Split(m.log.view, "\n")

	if maxLines > 0 {
		lines = lines[len(lines)-maxLines:]
	}

	var lastTimeStamp time.Time = time.Now()
	for i, line := range lines {
		if line == "" {
			continue
		}
		var err error
		var parts []string = strings.Split(line, " ")
		var tag string = parts[3]
		var subjectStyle lipgloss.Style
		var timeStamp time.Time

		timeStamp, err = time.Parse(logger.GetTimeFormat(), strings.Join(parts[:3], " "))
		if err != nil {
			logger.Log(logger.Error, err.Error())
		}

		timeDiff := timeStamp.Sub(lastTimeStamp)
		if i == 0 {
			timeDiff = time.Duration(0)
		}
		lastTimeStamp = timeStamp
		fmt.Fprintf(&timeBuilder, "%s \n", timeDiff.Round(10*time.Microsecond))

		tag, subjectStyle = getTagStyle(tag)
		fmt.Fprintf(&tagBuilder, "%s \n", tag)

		var words []string = parts[5:]

		bullet := subjectStyle.Render("<>")
		subj := fmt.Sprintf("%s %s", strings.Split(parts[4], ".")[0][1:], bullet)

		subjBuilder.WriteString(subj)
		subjBuilder.WriteString("\n ")
		fmt.Fprintf(&wordBuilder, " %s\n", slogWordStyle.Render(strings.Join(words, " ")))
	}

	content := lipgloss.JoinHorizontal(
		lipgloss.Left,
		slogPathStyle.Render(timeBuilder.String()),
		tagBuilder.String(),
		slogPathStyle.Render(subjBuilder.String()),
		wordBuilder.String(),
	)

	m.slog.lastRenderDur = time.Since(now)
	m.slog.view = content

	// The terminal can be resized at any time
	w, h := m.ViewportSize()
	m.slog.viewPort.SetWidth(w)
	m.slog.viewPort.SetHeight(h - 2)

	m.slog.viewPort.SetContent(m.slog.view)
	m.slog.viewPort.GotoBottom()
	return m, nil
}

func getTagStyle(tag string) (string, lipgloss.Style) {
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
	case "[∞∞∞]":
		tag = insaneLogStyle.Render(tag)
		subjectStyle = insaneLogStyle
	}

	return tag, subjectStyle
}

func clearSlogView(m debugModel) tea.View {
	w, h := m.ViewportSize()
	return tea.NewView(ui.NewInfoBox(
		"Clear Slog",
		"Slog has been reset and will be re-rendered on execution.",
		w, h,
	))
}

func viewSlog(m debugModel) tea.View {
	w, h := m.ViewportSize()
	m.slog.viewPort.SetWidth(w)
	m.slog.viewPort.SetHeight(h - 2)

	content := fmt.Sprintf(
		"%s\nTook: %s",
		m.slog.viewPort.View(),
		m.slog.lastRenderDur,
	)
	return tea.NewView(logStyle.Render(content))
}

func loadStats(m debugModel, arg *cmdcore.NoArg) (cmdcore.CommandModel, error) {
	var err error

	path := filepath.Join(utils.GetWorkingDir(), "log.txt")
	fileInfo, err := os.Stat(path)
	if err != nil {
		return m, err
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

	return m, nil
}

func viewStats(m debugModel) tea.View {
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

	w, h := m.ViewportSize()

	return tea.NewView(lipgloss.Place(
		w, h,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			lipgloss.JoinVertical(lipgloss.Left, debugHeader, debugStats, logHeader, logStats),
			" ",
			lipgloss.JoinVertical(lipgloss.Left, memHeader, memStats),
		),
	))
}

func setLogLevel(m debugModel, arg *logger.LogLevel) (cmdcore.CommandModel, error) {
	// We ignore error because this check is already done
	// in the parse func.
	_ = logger.SetLogLevel(*arg)
	return m, nil
}

func viewLogLevel(m debugModel) tea.View {
	w, h := m.ViewportSize()
	return tea.NewView(ui.NewInfoBox(
		"Set Log Level",
		fmt.Sprintf("Log level has been set to %s", logger.GetLogLevel()),
		w, h,
	))
}
