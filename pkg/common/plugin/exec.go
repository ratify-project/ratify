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
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	maxRetryCount = 5
	waitDuration  = time.Second
)

// Executor is an interface that defines methods to lookup a plugin and execute it.
type Executor interface {
	// ExecutePlugin executes the plugin with the given parameters
	ExecutePlugin(ctx context.Context, pluginPath string, cmdArgs []string, stdinData []byte, environ []string) ([]byte, error)
	// FindInPaths finds the plugin in the given paths
	FindInPaths(plugin string, paths []string) (string, error)
}

// DefaultExecutor finds the plugin executable and invokes it as a os command
type DefaultExecutor struct {
	Stderr io.Writer
}

// return the command output and the error
func (e *DefaultExecutor) ExecutePlugin(ctx context.Context, pluginPath string, cmdArgs []string, stdinData []byte, environ []string) ([]byte, error) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	c := exec.CommandContext(ctx, pluginPath, cmdArgs...)
	c.Env = environ
	c.Stdin = bytes.NewBuffer(stdinData)
	c.Stdout = stdout
	c.Stderr = stderr

	// DEBUG: log the process details used to launch the binary plugin
	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		logrus.Debugf("launching plugin %s", pluginPath)

		pluginEnv := make([]string, 3)
		for _, env := range environ {
			// plugins inherit all env vars, but we're interested in the RATIFY_* ones. This also helps to keep secret values out of the logs.
			if strings.HasPrefix(env, "RATIFY_") {
				pluginEnv = append(pluginEnv, env)
			}
		}
		logrus.Debugf("env vars: %v", pluginEnv)
		logrus.Debugf("args: %v", cmdArgs)
		logrus.Debugf("stdin: %s", stdinData)
	}

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
			logrus.Debugf("command returned text file busy, retrying after %v", waitDuration)
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
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()

			switch {
			case strings.HasPrefix(strings.ToLower(line), "info"):
				line = strings.Replace(line, "info: ", "", -1)
				logrus.Infof("[Plugin] %s", line)
			case strings.HasPrefix(strings.ToLower(line), "warn"):
				line = strings.Replace(line, "warn: ", "", -1)
				logrus.Warnf("[Plugin] %s", line)
			case strings.HasPrefix(strings.ToLower(line), "debug"):
				line = strings.Replace(line, "debug: ", "", -1)
				logrus.Debugf("[Plugin] %s", line)
			default:
				fmt.Fprintf(os.Stderr, "[Plugin] %s,", line)
			}
		}
		// TODO: Should this be removed since the msgs are processed?
		//_, _ = stderr.WriteTo(e.Stderr)
	}

	var obj interface{}
	var resultsBuffer bytes.Buffer

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "{") {
			err := json.NewDecoder(strings.NewReader(line)).Decode(&obj)
			if err != nil {
				continue
			}

			jsonString, _ := json.Marshal(obj)
			resultsBuffer.WriteString(string(jsonString))
		} else {
			fmt.Printf("[Plugin] %s \n", line)
		}
	}

	// TODO stdout reader
	return resultsBuffer.Bytes(), nil
}

func (e *DefaultExecutor) pluginErr(err error, stdout, stderr []byte) error {
	var stdOutBuffer bytes.Buffer
	var stdErrBuffer bytes.Buffer

	// writing them to buffer to avoid lint error
	stdOutBuffer.Write(stdout)
	stdErrBuffer.Write(stderr)

	errCombined := Error{}
	errCombined.Msg = fmt.Sprintf("plugin failed with error: '%v', msg from stError '%v', msg from stdOut '%v'", err.Error(), stdErrBuffer.String(), stdOutBuffer.String())
	return &errCombined
}

func (e *DefaultExecutor) FindInPaths(plugin string, paths []string) (string, error) {
	return FindInPaths(plugin, paths)
}
