package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

type Logger struct {
	infoLogger  *log.Logger
	debugLogger *log.Logger
	warnLogger  *log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		infoLogger:  log.New(os.Stderr, "INFO: ", 0),
		debugLogger: log.New(os.Stderr, "DEBUG: ", 0),
		warnLogger:  log.New(os.Stderr, "WARN: ", 0),
	}
}

func (l *Logger) SetOutput(out io.Writer) {
	l.infoLogger.SetOutput(out)
	l.debugLogger.SetOutput(out)
	l.warnLogger.SetOutput(out)
}

func (l *Logger) Info(message string) {
	l.infoLogger.Println(message)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.infoLogger.Println(message)
}

func (l *Logger) Debug(message string) {
	l.debugLogger.Println(message)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.debugLogger.Println(message)
}

func (l *Logger) Warn(message string) {
	l.warnLogger.Println(message)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.warnLogger.Println(message)
}
