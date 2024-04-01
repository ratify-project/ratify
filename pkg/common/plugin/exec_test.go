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
	"bytes"
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

func TestParsePluginOutput_EmptyBuffers(t *testing.T) {
	stdOut := bytes.NewBufferString("")
	stdErr := bytes.NewBufferString("")

	json, messages := parsePluginOutput(stdOut, stdErr)

	if len(messages) != 0 {
		t.Fatalf("unexpected messages, expected 0, got %d", len(messages))
	}

	if len(json) != 0 {
		t.Fatalf("unexpected json, expected empty, got '%s'", json)
	}
}

func TestParsePluginOutput_StdErrNonEmpty(t *testing.T) {
	stdOut := bytes.NewBufferString("")
	stdErr := bytes.NewBufferString("This is a string from std err\n")

	json, messages := parsePluginOutput(stdOut, stdErr)

	if len(messages) != 1 {
		t.Fatalf("unexpected messages, expected 1, got %d", len(messages))
	}

	if len(json) != 0 {
		t.Fatalf("unexpected json, expected empty, got '%s'", json)
	}
}

func TestParsePluginOutput_StdOutNonEmpty(t *testing.T) {
	stdOut := bytes.NewBufferString("{\"key\":\"value\"}\n")
	stdErr := bytes.NewBufferString("")

	expectedJSON := []byte(`{"key":"value"}`)

	json, messages := parsePluginOutput(stdOut, stdErr)

	if len(messages) != 0 {
		t.Fatalf("unexpected messages, expected 0, got %d", len(messages))
	}

	if !bytes.Equal(expectedJSON, json) {
		t.Fatalf("unexpected json, expected '%s', got '%s'", expectedJSON, json)
	}
}

func TestParsePluginOutput_InvalidJson(t *testing.T) {
	stdOut := bytes.NewBufferString("This is not a valid json\n")
	stdErr := bytes.NewBufferString("")

	json, messages := parsePluginOutput(stdOut, stdErr)

	if len(messages) != 1 {
		t.Fatalf("unexpected messages, expected 1, got %d", len(messages))
	}

	if len(json) != 0 {
		t.Fatalf("unexpected json, expected empty, got '%s'", json)
	}
}
