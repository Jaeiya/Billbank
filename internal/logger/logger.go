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

	"charm.land/lipgloss/v2"
)

var (
	_logChan         chan LogMsg
	_doneChan        chan struct{}
	_isReady         = false
	_logLevel        = None
	_stdLogger       *log.Logger
	_fileHandle      *os.File
	_timeFormat      = "03:04:05.000000 PM MST"
	_defaultFileName = "log.txt"
	_filePath        = ""
	_mux             sync.Mutex
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
		return false
	}
	return true
}

type LogMsg struct {
	msg   func() string
	level LogLevel
	file  string
	line  int
	vars  []any
}

func Log(ll LogLevel, msg string, vars ...any) {
	if _logLevel == None || ll < _logLevel {
		return
	}

	tryInitLog()

	_, file, line, _ := runtime.Caller(1)
	_logChan <- LogMsg{func() string { return msg }, ll, file, line, vars}
}

// LogFunc executes the msgFn and passes its result to the
// default log func, if the specified log level is active.
//
// This is useful if a log requires some heavier processing
// but you don't want it to affect the runtime of your
// application. The processing will happen inside the
// logger thread instead.
func LogFunc(ll LogLevel, msgFn func() string, vars ...any) {
	if _logLevel == None || ll < _logLevel {
		return
	}

	tryInitLog()

	if ll < _logLevel {
		return
	}

	_, file, line, _ := runtime.Caller(1)
	_logChan <- LogMsg{msgFn, ll, file, line, vars}
}

func LogFatal(errMsg string, description string, vars ...any) {
	defer func() {
		err := CloseLog()
		if err != nil {
			// Should effectively never happen
			panic(err)
		}
		os.Exit(1)
	}()

	maxDisplayWidth := 60

	errMsgStyle := lipgloss.NewStyle().
		Width(maxDisplayWidth).
		PaddingTop(1).
		PaddingLeft(1).
		Foreground(lipgloss.Red)

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
			Foreground(lipgloss.Yellow).
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

func Reset() error {
	_mux.Lock()
	ll := _logLevel
	// Do not allow any calls to log during reset
	_logLevel = None

	// Init should be called once reset is done
	_isReady = false
	_mux.Unlock()

	err := CloseLog()
	if err != nil {
		return err
	}
	_logLevel = ll
	return nil
}

func SetLogLevel(ll LogLevel) error {
	if !ll.IsValid() {
		return fmt.Errorf("invalid log level::%d", ll)
	}
	_logLevel = ll
	return nil
}

func GetLogLevel() LogLevel {
	return _logLevel
}

func GetTimeFormat() string {
	return _timeFormat
}

func SetFilePath(path string) {
	_filePath = path
}

func CloseLog() error {
	close(_logChan)
	<-_doneChan
	return _fileHandle.Close()
}

func tryInitLog() {
	if _isReady {
		return
	}

	if _filePath == "" {
		wd, err := os.Getwd()
		if err != nil {
			// This should effectively never happen
			panic(err)
		}
		_filePath = filepath.Join(wd, _defaultFileName)
	}

	_logChan = make(chan LogMsg, 50)
	_doneChan = make(chan struct{})

	var err error
	_fileHandle, err = os.OpenFile(
		_filePath,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0o644,
	)
	if err != nil {
		panic(err)
	}

	_stdLogger = log.New(_fileHandle, "", 0)
	go logMessages()
	_isReady = true
}

func logMessages() {
	for log := range _logChan {
		msg := fmt.Sprintf(
			"%s [%s] [%s:%d]: %s\n",
			time.Now().Format(_timeFormat),
			log.level,
			filepath.Base(log.file), log.line,
			log.msg(),
		)
		if len(log.vars) == 0 {
			_stdLogger.Print(msg)
		} else {
			_stdLogger.Printf(msg, log.vars...)
		}
	}
	close(_doneChan)
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
		fmt.Fprintf(&funcBuilder, "\t%s(): \n", function)
		fmt.Fprintf(&fileBuilder, "%s:%d\n", file, frame.Line)
		if !more {
			break
		}
	}
	funcStyle := lipgloss.NewStyle().
		Foreground(lipgloss.BrightBlack).
		Align(lipgloss.Right).
		Render(funcBuilder.String())
	display := lipgloss.JoinHorizontal(lipgloss.Top, funcStyle, fileBuilder.String())
	return display
}
