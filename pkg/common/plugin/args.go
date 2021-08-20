package plugin

import (
	"fmt"
	"strings"
)

type PluginArgs interface {
	// For use with os/exec; i.e., return nil to inherit the
	// environment from this process
	AsEnv() []string
}

type inherited struct{}

var inheritArgsFromEnv inherited

func (*inherited) AsEnv() []string {
	return nil
}

func ArgsFromEnv() PluginArgs {
	return &inheritArgsFromEnv
}

func Stringify(pluginArgs [][2]string) string {
	entries := make([]string, len(pluginArgs))

	for i, kv := range pluginArgs {
		entries[i] = strings.Join(kv[:], "=")
	}

	return strings.Join(entries, ";")
}

// dedupEnv returns a copy of env with any duplicates removed, in favor of later values.
// Items not of the normal environment "key=value" form are preserved unchanged.
func DedupEnv(env []string) []string {
	out := make([]string, 0, len(env))
	envMap := map[string]string{}

	for _, kv := range env {
		// find the first "=" in environment, if not, just keep it
		eq := strings.Index(kv, "=")
		if eq < 0 {
			out = append(out, kv)
			continue
		}
		envMap[kv[:eq]] = kv[eq+1:]
	}

	for k, v := range envMap {
		out = append(out, fmt.Sprintf("%s=%s", k, v))
	}

	return out
}

func ParseArgs(args string) ([][2]string, error) {

	if args == "" {
		return nil, nil
	}

	var pluginArgs [][2]string

	pairs := strings.Split(args, ";")
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) != 2 {
			return nil, fmt.Errorf("ARGS: invalid kv pair %q", pair)
		}
		pluginArgs = append(pluginArgs, [2]string{kv[0], kv[1]})
	}
	return pluginArgs, nil
}
