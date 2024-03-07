/*
Copyright The Ratify Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
