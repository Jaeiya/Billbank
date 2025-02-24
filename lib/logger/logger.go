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

const (
	Insane = LogLevel(iota)
	Hot
	Debug
	Info
	Attention
	Error
	None
)

var levelTags = []string{
	"∞∞∞", "HOT", "DBG", "NFO", "ATN", "ERR", "None",
}

type LogLevel int

func (ll LogLevel) String() string {
	return levelTags[ll]
}

func (ll LogLevel) IsValid() bool {
	if ll < Insane || ll > None {
		LogFatal("log level not found [%d]", "This should not happen!", ll)
	}
	return true
}

type LogMsg struct {
	msg   string
	level LogLevel
	file  string
	line  int
	vars  []any
}

var (
	isReady    = false
	logLevel   = None
	logChan    = make(chan LogMsg, 50)
	doneChan   = make(chan struct{})
	stdLogger  *log.Logger
	fileHandle *os.File
	once       sync.Once
)

func Log(ll LogLevel, msg string, vars ...any) {
	// Removes side-effects when testing code that is
	// using the logger.
	if logLevel == None {
		return
	}

	if ll.IsValid() {
		initLog()

		if ll < logLevel {
			return
		}

		_, file, line, _ := runtime.Caller(1)
		logChan <- LogMsg{msg, ll, file, line, vars}
	}
}

// LogFunc executes the msgFn and passes its result to the
// default log func, if the specified log level is active.
//
// This is useful if you need a log that does some heavy
// processing, but only want that processing to occur at
// a specific log level.
func LogFunc(ll LogLevel, msgFn func() string, vars ...any) {
	// Removes side-effects when testing code that is
	// using the logger.
	if logLevel == None {
		return
	}

	if ll < logLevel {
		return
	}

	if ll.IsValid() {
		initLog()

		if ll < logLevel {
			return
		}

		_, file, line, _ := runtime.Caller(1)
		logChan <- LogMsg{msgFn(), ll, file, line, vars}
	}
}

func LogFatal(errMsg string, description string, vars ...any) {
	defer func() {
		CloseLog()
		os.Exit(1)
	}()

	maxDisplayWidth := 60

	errMsgStyle := lipgloss.NewStyle().
		Width(maxDisplayWidth).
		PaddingTop(1).
		PaddingLeft(1).
		Foreground(ui.Red)

	description = strings.TrimSpace(description)
	if len(description) > 0 {
		paragraphs := strings.Split(description, "\n\n")
		for i, p := range paragraphs {
			paragraphs[i] = lipgloss.NewStyle().
				PaddingTop(1).
				Width(maxDisplayWidth).
				Render(strings.ReplaceAll(p, "\n", " "))
		}
		description = lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingBottom(1).
			Foreground(ui.Yellow).
			Render(lipgloss.JoinVertical(lipgloss.Left, paragraphs...))
	}

	if len(vars) > 0 {
		errMsg = errMsgStyle.Render(fmt.Sprintf(errMsg, vars...))
	} else {
		errMsg = errMsgStyle.Render(errMsg)
	}

	fmt.Print(
		lipgloss.JoinVertical(
			lipgloss.Left,
			errMsg,
			description,
			getStack(),
		) + "\n",
	)
}

func SetLogLevel(ll LogLevel) {
	if ll.IsValid() {
		logLevel = ll
	}
}

func GetLogLevel() LogLevel {
	return logLevel
}

func CloseLog() error {
	close(logChan)
	<-doneChan
	return fileHandle.Close()
}

func initLog() {
	once.Do(func() {
		if isReady {
			return
		}
		path := filepath.Join(utils.GetWorkingDir(), "log.txt")

		var err error
		fileHandle, err = os.OpenFile(
			path,
			os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
			0o644,
		)
		if err != nil {
			panic(err)
		}

		stdLogger = log.New(fileHandle, "", 0)
		go logMessages()
		isReady = true
	})
}

func logMessages() {
	for log := range logChan {
		msg := fmt.Sprintf(
			"%s [%s] [%s:%d]: %s\n",
			time.Now().Format("03:04:05.000 PM MST"),
			log.level,
			filepath.Base(log.file), log.line,
			log.msg,
		)
		if len(log.vars) == 0 {
			stdLogger.Print(msg)
		} else {
			stdLogger.Printf(msg, log.vars...)
		}
	}
	close(doneChan)
}

func getStack() string {
	pc := make([]uintptr, 4)
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
