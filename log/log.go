package log

import (
	"os"

	"github.com/charmbracelet/log"
)

var (
	debug  = os.Getenv("DEBUG") != ""
	logger log.Logger
)

func init() {
	if debug {
		var (
			logFile *os.File
		)
		filename := "log.txt"
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			logFile, _ = os.Create(filename)
		} else {
			logFile, _ = os.OpenFile(filename, os.O_RDWR|os.O_APPEND, 0660)
		}
		logger = log.New(log.WithOutput(logFile), log.WithLevel(log.DebugLevel))
	} else {
		logger = log.New(log.WithLevel(log.InfoLevel))
	}
}

func Debug(msg any, keyval ...any) {
	logger.Debug(msg, keyval...)
}

func Fatal(msg any, keyval ...any) {
	logger.Fatal(msg, keyval...)
}
