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
