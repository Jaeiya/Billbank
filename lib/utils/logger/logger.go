package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/jaeiya/billbank/lib/utils"
)

type LogLevel int

const (
	Hot = LogLevel(iota)
	Debug
	Info
	Attention
	Error
)

type LogMsg struct {
	msg     string
	level   LogLevel
	file    string
	line    int
	subject string
	vars    []any
}

var (
	isReady  = false
	logger   *log.Logger
	logChan  = make(chan LogMsg, 50)
	logLevel LogLevel
)

func NewLog(ll LogLevel) bool {
	if isReady {
		return true
	}

	path := filepath.Join(utils.GetWorkingDir(), "log.txt")

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC|os.O_APPEND, 0o644)
	if err != nil {
		panic(err)
	}

	logLevel = ll
	logger = log.New(file, "", 0)
	go logMessages()
	isReady = true

	return true
}

func Log(ll LogLevel, subject string, msg string, vars ...any) {
	if ll < logLevel {
		return
	}

	if !isReady {
		panic("log not initialized")
	}

	_, file, line, _ := runtime.Caller(1)
	logChan <- LogMsg{msg, ll, file, line, subject, vars}
}

func SetLogLevel(ll LogLevel) {
	logLevel = ll
}

func CloseLog() {
	close(logChan)
}

func logMessages() {
	for log := range logChan {
		msg := fmt.Sprintf(
			"%s [%s] [%s:%d]: %s: %s\n",
			time.Now().Format("03:04:05.000 PM MST"),
			getLogLevelStr(log.level),
			filepath.Base(log.file), log.line,
			log.subject,
			log.msg,
		)
		if len(log.vars) == 0 {
			logger.Print(msg)
		} else {
			logger.Printf(msg, log.vars...)
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
	default:
		panic("invalid log level")
	}
}
