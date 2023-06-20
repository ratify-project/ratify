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
	"fmt"
	"strings"
)

type PluginArgs interface { //nolint:revive // ignore linter to have unique type name
	AsEnviron() []string
}

// Concat joins the given array of key=value formatted strings into a single ; delimited string.
func Concat(pluginArgs [][2]string) string {
	entries := make([]string, len(pluginArgs))

	for i, kv := range pluginArgs {
		entries[i] = strings.Join(kv[:], "=")
	}

	return strings.Join(entries, ";")
}

// MergeDuplicateEnviron returns a copy of environment variables with any duplicates removed, and keeping the latest values.
// Only variables of format "key=value" are considered for merging.
func MergeDuplicateEnviron(env []string) []string {
	out := make([]string, 0, len(env))
	envMap := map[string]string{}

	for _, kv := range env {
		// find the first "=" in environment variable, if not, skip it
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

// ParseInputArgs parses the given string into an array of key=value formatted strings.
func ParseInputArgs(args string) ([][2]string, error) {
	if args == "" {
		return nil, nil
	}

	pluginArgs := [][2]string{}
	pairs := strings.Split(args, ";")
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) != 2 {
			return nil, fmt.Errorf("PLUGIN ARGS: invalid kv pair %q", pair)
		}
		pluginArgs = append(pluginArgs, [2]string{kv[0], kv[1]})
	}
	return pluginArgs, nil
}
