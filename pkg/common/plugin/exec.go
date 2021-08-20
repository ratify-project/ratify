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

type Exec interface {
	ExecPlugin(ctx context.Context, pluginPath string, cmdArgs []string, stdinData []byte, environ []string) ([]byte, error)
	FindInPath(plugin string, paths []string) (string, error)
}

type DefaultExec struct {
	Stderr io.Writer
}

func (e *DefaultExec) ExecPlugin(ctx context.Context, pluginPath string, cmdArgs []string, stdinData []byte, environ []string) ([]byte, error) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	c := exec.CommandContext(ctx, pluginPath, cmdArgs...)
	c.Env = environ
	c.Stdin = bytes.NewBuffer(stdinData)
	c.Stdout = stdout
	c.Stderr = stderr

	// Retry the command on "text file busy" errors
	for i := 0; i <= 5; i++ {
		err := c.Run()

		// Command succeeded
		if err == nil {
			break
		}

		// If the plugin is currently about to be written, then we wait a
		// second and try it again
		if strings.Contains(err.Error(), "text file busy") {
			time.Sleep(time.Second)
			continue
		}

		// All other errors except than the busy text file
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

func (e *DefaultExec) pluginErr(err error, stdout, stderr []byte) error {
	emsg := Error{}
	if len(stdout) == 0 {
		if len(stderr) == 0 {
			emsg.Msg = fmt.Sprintf("plugin failed with no error message: %v", err)
		} else {
			emsg.Msg = fmt.Sprintf("plugin failed: %q", string(stderr))
		}
	} else if perr := json.Unmarshal(stdout, &emsg); perr != nil {
		emsg.Msg = fmt.Sprintf("plugin failed but error parsing its diagnostic message %q: %v", string(stdout), perr)
	}
	return &emsg
}

func (e *DefaultExec) FindInPath(plugin string, paths []string) (string, error) {
	return FindInPath(plugin, paths)
}
