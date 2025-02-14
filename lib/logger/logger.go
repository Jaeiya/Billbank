package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/jaeiya/billbank/lib/ui"
	"github.com/jaeiya/billbank/lib/utils"
)

type LogLevel int

const (
	Insane = LogLevel(iota)
	Hot
	Debug
	Info
	Attention
	Error
	None
)

type LogMsg struct {
	msg   string
	level LogLevel
	file  string
	line  int
	vars  []any
}

var (
	isReady   = false
	logLevel  = None
	logChan   = make(chan LogMsg, 50)
	stdLogger *log.Logger
	once      sync.Once
)

func Log(ll LogLevel, msg string, vars ...any) {
	// Removes side-effects when testing code that is
	// using the logger.
	if logLevel == None {
		return
	}

	initLog()

	if ll < logLevel {
		return
	}

	_, file, line, _ := runtime.Caller(1)
	logChan <- LogMsg{msg, ll, file, line, vars}
}

func LogFatal(errMsg string, description string, vars ...any) {
	defer func() {
		CloseLog()
		os.Exit(1)
	}()

	errMsg = lipgloss.NewStyle().
		Width(60).
		PaddingTop(1).
		PaddingLeft(1).
		Foreground(ui.Red).
		Render(errMsg)

	description = strings.TrimSpace(description)
	description = fmt.Sprintf("\n%s\n", strings.ReplaceAll(description, "\n", " "))
	description = lipgloss.NewStyle().
		PaddingLeft(1).
		Width(60).
		Foreground(ui.Yellow).
		Render(description)

	if len(vars) > 0 {
		fmt.Printf(errMsg, vars...)
		fmt.Print(description)
		fmt.Printf("\n%s\n", getStack())
	} else {
		fmt.Println(errMsg)
		fmt.Println(description)
	}
}

func SetLogLevel(ll LogLevel) {
	logLevel = ll
}

func CloseLog() {
	close(logChan)
}

func initLog() {
	once.Do(func() {
		if isReady {
			return
		}
		path := filepath.Join(utils.GetWorkingDir(), "log.txt")

		file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC|os.O_APPEND, 0o644)
		if err != nil {
			panic(err)
		}

		stdLogger = log.New(file, "", 0)
		go logMessages()
		isReady = true
	})
}

func logMessages() {
	for log := range logChan {
		msg := fmt.Sprintf(
			"%s [%s] [%s:%d]: %s\n",
			time.Now().Format("03:04:05.000 PM MST"),
			getLogLevelStr(log.level),
			filepath.Base(log.file), log.line,
			log.msg,
		)
		if len(log.vars) == 0 {
			stdLogger.Print(msg)
		} else {
			stdLogger.Printf(msg, log.vars...)
		}
	}
}

func getLogLevelStr(ll LogLevel) string {
	switch ll {
	case Info:
		return "NFO"
	case Attention:
		return "ATN"
	case Error:
		return "ERR"
	case Debug:
		return "DBG"
	case Hot:
		return "HOT"
	case Insane:
		return "∞∞∞"
	default:
		panic("invalid log level")
	}
}

func getStack() string {
	pc := make([]uintptr, 3)
	n := runtime.Callers(3, pc)
	if n == 0 {
		return ""
	}

	pc = pc[:n]
	frames := runtime.CallersFrames(pc)

	// var sb strings.Builder
	var funcBuilder strings.Builder
	var fileBuilder strings.Builder
	for {
		frame, more := frames.Next()
		file := filepath.Base(frame.File)
		function := filepath.Base(frame.Function)
		funcBuilder.WriteString(fmt.Sprintf("\t%s(): \n", function))
		fileBuilder.WriteString(fmt.Sprintf("%s:%d\n", file, frame.Line))
		if !more {
			break
		}
	}
	funcStyle := lipgloss.NewStyle().
		Foreground(ui.Gray).
		Align(lipgloss.Right).
		Render(funcBuilder.String())
	display := lipgloss.JoinHorizontal(lipgloss.Top, funcStyle, fileBuilder.String())
	return display
}
