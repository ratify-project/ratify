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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

const (
	maxRetryCount = 5
	waitDuration  = time.Second
)

type Executor interface {
	ExecutePlugin(ctx context.Context, pluginPath string, cmdArgs []string, stdinData []byte, environ []string) ([]byte, error)
	FindInPaths(plugin string, paths []string) (string, error)
}

type DefaultExecutor struct {
	Stderr io.Writer
}

func (e *DefaultExecutor) ExecutePlugin(ctx context.Context, pluginPath string, cmdArgs []string, stdinData []byte, environ []string) ([]byte, error) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	c := exec.CommandContext(ctx, pluginPath, cmdArgs...)
	c.Env = environ
	c.Stdin = bytes.NewBuffer(stdinData)
	c.Stdout = stdout
	c.Stderr = stderr

	// Retry the command on "text file busy" errors
	for i := 0; i <= maxRetryCount; i++ {
		err := c.Run()

		// Command succeeded
		if err == nil {
			break
		}

		// If the plugin is about to be completed, then we wait a
		// second and try it again
		if strings.Contains(err.Error(), "text file busy") {
			time.Sleep(waitDuration)
			continue
		}

		// For all other errors return failed.
		return nil, e.pluginErr(err, stdout.Bytes(), stderr.Bytes())
	}

	// Copy stderr to caller's buffer in case plugin printed to both
	// stdout and stderr for some reason. Ignore failures as stderr is
	// only informational.
	if e.Stderr != nil && stderr.Len() > 0 {
		_, _ = stderr.WriteTo(e.Stderr)
	}
	// TODO stdout reader
	return stdout.Bytes(), nil
}

func (e *DefaultExecutor) pluginErr(err error, stdout, stderr []byte) error {
	errMsg := Error{}
	if len(stdout) == 0 {
		if len(stderr) == 0 {
			errMsg.Msg = fmt.Sprintf("plugin failed with no proper error message: %v", err)
		} else {
			errMsg.Msg = fmt.Sprintf("plugin failed with error: %q", string(stderr))
		}
	} else if perr := json.Unmarshal(stdout, &errMsg); perr != nil {
		errMsg.Msg = fmt.Sprintf("plugin failed and parsing its error message also failed with error %q: %v", string(stdout), perr)
	}
	return &errMsg
}

func (e *DefaultExecutor) FindInPaths(plugin string, paths []string) (string, error) {
	return FindInPaths(plugin, paths)
}
