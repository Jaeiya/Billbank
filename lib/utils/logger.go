package utils

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type LogLevel int

const (
	Debug = -1
	Info  = LogLevel(iota)
	Attention
	Error
)

type LogMsg struct {
	msg   any
	level LogLevel
	file  string
	line  int
}

var (
	isReady = false
	logger  *log.Logger
	logChan = make(chan LogMsg, 50)
)

func Log(ll LogLevel, msg any) {
	if !isReady {
		panic("log not initialized")
	}
	_, file, line, _ := runtime.Caller(1)
	logChan <- LogMsg{msg, ll, file, line}
}

func CloseLog() {
	close(logChan)
}

func CreateLog() bool {
	if isReady {
		return true
	}

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	path := filepath.Join(wd, "log.txt")

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC|os.O_APPEND, 0o644)
	if err != nil {
		panic(err)
	}

	logger = log.New(file, "", 0)
	go logMessages()
	isReady = true

	return true
}

func logMessages() {
	for msg := range logChan {
		logger.Printf(
			"%s [%s] [%s:%d]: %v\n",
			time.Now().Format("03:04:05 PM MST"),
			getLogLevelStr(msg.level),
			filepath.Base(msg.file), msg.line,
			msg.msg,
		)
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
	default:
		panic("invalid log level")
	}
}
