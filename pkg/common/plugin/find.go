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
	"os"
	"path/filepath"
)

// FindInPaths returns the full path of the plugin executable by searching in the provided list of paths
func FindInPaths(plugin string, paths []string) (string, error) {
	if plugin == "" {
		return "", fmt.Errorf("plugin name is required")
	}

	if len(paths) == 0 {
		return "", fmt.Errorf("no paths provided to find a plugin")
	}

	for _, path := range paths {
		for _, fe := range executableFileExtensions {
			fullpath := filepath.Join(path, plugin) + fe
			if fi, err := os.Stat(fullpath); err == nil && fi.Mode().IsRegular() {
				return fullpath, nil
			}
		}
	}

	return "", fmt.Errorf("failed to find plugin %q in paths %s", plugin, paths)
}
