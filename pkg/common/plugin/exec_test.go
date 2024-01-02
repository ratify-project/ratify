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

package plugin

import (
	"strings"
	"testing"
)

func TestPluginErr(t *testing.T) {
	stdOut := []byte("This is a string from std out")
	stdErr := []byte("This is a string from std err")
	errMsg := Error{}
	errMsg.Msg = "plugin error"

	e := DefaultExecutor{}
	err := e.pluginErr(&errMsg, stdOut, stdErr)
	if err == nil {
		t.Fatalf("plugin error expected")
	}

	if !strings.Contains(err.Error(), errMsg.Msg) {
		t.Fatalf("error msg should contain stdOut error msg, actual '%v'", err.Error())
	}

	if !strings.Contains(err.Error(), string(stdOut)) {
		t.Fatalf("error msg should contain stdOut msg, actual '%v'", err.Error())
	}

	if !strings.Contains(err.Error(), string(stdErr)) {
		t.Fatalf("error msg should contain stdErr msg, actual '%v'", err.Error())
	}
}
