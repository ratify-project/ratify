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
	"testing"

	"github.com/sirupsen/logrus"
)

func TestSetLoggingLevel_DefaultsToInfo_NoEnvVar(t *testing.T) {
	SetLoggingLevel("")

	if logrus.GetLevel() != logrus.InfoLevel {
		t.Errorf("Expected log level to be %s, got %s", logrus.InfoLevel, logrus.GetLevel())
	}
}

func TestSetLoggingLevel_DefaultsToInfo_BadEnvVar(t *testing.T) {
	SetLoggingLevel("undefinedlogginglevel")

	if logrus.GetLevel() != logrus.InfoLevel {
		t.Errorf("Expected log level to be %s, got %s", logrus.InfoLevel, logrus.GetLevel())
	}
}

func TestSetLoggingLevel_UsesEnvVar(t *testing.T) {
	SetLoggingLevel("debug")

	if logrus.GetLevel() != logrus.DebugLevel {
		t.Errorf("Expected log level to be %s, got %s", logrus.DebugLevel, logrus.GetLevel())
	}
}
