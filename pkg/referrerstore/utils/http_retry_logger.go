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

package utils

import (
	"github.com/hashicorp/go-retryablehttp"
	"github.com/sirupsen/logrus"
)

// NOTE: This Logger is an adapter for logrus used by Retryable HTTP client in ORAS
type HttpRetryLogger struct {
	retryablehttp.LeveledLogger
}

func (l HttpRetryLogger) Error(msg string, keysAndValues ...interface{}) {
	logrus.Errorln(append([]interface{}{msg}, keysAndValues...)...)
}

func (l HttpRetryLogger) Info(msg string, keysAndValues ...interface{}) {
	logrus.Infoln(append([]interface{}{msg}, keysAndValues...)...)
}

func (l HttpRetryLogger) Debug(msg string, keysAndValues ...interface{}) {
	logrus.Debugln(append([]interface{}{msg}, keysAndValues...)...)
}

func (l HttpRetryLogger) Warn(msg string, keysAndValues ...interface{}) {
	logrus.Warnln(append([]interface{}{msg}, keysAndValues...)...)
}
