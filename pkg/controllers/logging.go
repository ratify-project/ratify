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

package controllers

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
)

// LogrusSink is an adapter to allow the use of logrus with logr, as required by k8s controller-runtime.
type LogrusSink struct {
	logger        *logrus.Logger
	names         []string
	keysAndValues []interface{}
}

func NewLogrusSink(logger *logrus.Logger) *LogrusSink {
	return &LogrusSink{
		logger: logger,
	}
}

// Init receives optional information about the logr library for LogSink
// implementations that need it.
func (sink *LogrusSink) Init(_ logr.RuntimeInfo) {
}

// Enabled tests whether this LogSink is enabled at the specified V-level.
// For example, commandline flags might be used to set the logging
// verbosity and disable some info logs.
func (sink *LogrusSink) Enabled(level int) bool {
	// controller-runtime uses V=0 for Info and V=1 for Debug: https://github.com/kubernetes-sigs/controller-runtime/blob/master/TMP-LOGGING.md
	switch level {
	case 0:
		return sink.logger.IsLevelEnabled(logrus.InfoLevel)
	default:
		return sink.logger.IsLevelEnabled(logrus.DebugLevel)
	}
}

// Info logs a non-error message with the given key/value pairs as context.
// The level argument is provided for optional logging.  This method will
// only be called when Enabled(level) is true. See Logger.Info for more
// details.
func (sink *LogrusSink) Info(_ int, msg string, keysAndValues ...interface{}) {
	entry := sink.createEntry(keysAndValues...)
	entry.Info(sink.formatMessage(msg))
}

// Error logs an error, with the given message and key/value pairs as
// context.  See Logger.Error for more details.
func (sink *LogrusSink) Error(err error, msg string, keysAndValues ...interface{}) {
	entry := sink.createEntry(keysAndValues...)

	if sink.logger.IsLevelEnabled(logrus.DebugLevel) {
		stacktrace := string(debug.Stack())
		entry = entry.WithField("stacktrace", stacktrace)
	}

	entry.WithError(err).Error(sink.formatMessage(msg))
}

// WithValues returns a new LogSink with additional key/value pairs.  See
// Logger.WithValues for more details.
func (sink *LogrusSink) WithValues(keysAndValues ...interface{}) logr.LogSink {
	newSink := &LogrusSink{
		logger:        sink.logger,
		names:         sink.names,
		keysAndValues: append(sink.keysAndValues, keysAndValues...),
	}
	return newSink
}

// WithName returns a new LogSink with the specified name appended.  See
// Logger.WithName for more details.
func (sink *LogrusSink) WithName(name string) logr.LogSink {
	newSink := &LogrusSink{
		logger:        sink.logger,
		names:         append(sink.names, name),
		keysAndValues: sink.keysAndValues,
	}
	return newSink
}

func (sink *LogrusSink) createEntry(keysAndValues ...interface{}) *logrus.Entry {
	entry := logrus.NewEntry(sink.logger)

	if sink.names != nil {
		entry = entry.WithField("name", strings.Join(sink.names, "."))
	}

	if sink.keysAndValues != nil {
		for i := 0; i < len(sink.keysAndValues); i += 2 {
			entry = entry.WithField(fmt.Sprintf("%v", sink.keysAndValues[i]), sink.keysAndValues[i+1])
		}
	}

	if keysAndValues != nil {
		for i := 0; i < len(keysAndValues); i += 2 {
			entry = entry.WithField(fmt.Sprintf("%v", keysAndValues[i]), keysAndValues[i+1])
		}
	}

	return entry
}

func (sink *LogrusSink) formatMessage(msg string) string {
	if sink.names == nil || len(sink.names) == 0 {
		return msg
	}

	return fmt.Sprintf("[%s] %s", strings.Join(sink.names, "."), msg)
}
