package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

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
	Fatal
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
