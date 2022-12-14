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

package common

import (
	"github.com/sirupsen/logrus"
)

const defaultLevel = logrus.InfoLevel

func SetLoggingLevel(level string) {
	var logrusLevel logrus.Level
	if level == "" {
		logrusLevel = defaultLevel
	} else {
		var err error
		logrusLevel, err = logrus.ParseLevel(level)
		if err != nil {
			logrusLevel = defaultLevel
			logrus.Infof("Invalid log level %s, defaulting to %s", level, defaultLevel)
			logrus.Infof("Valid log levels are: %v", logrus.AllLevels)
		}
	}
	logrus.Infof("Setting log level to %s", logrusLevel)
	logrus.SetLevel(logrusLevel)
}
