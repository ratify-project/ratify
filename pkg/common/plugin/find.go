package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FindInPath returns the full path of the plugin by searching in the provided path
func FindInPath(plugin string, paths []string) (string, error) {
	if plugin == "" {
		return "", fmt.Errorf("no plugin name provided")
	}

	if strings.ContainsRune(plugin, os.PathSeparator) {
		return "", fmt.Errorf("invalid plugin name: %s", plugin)
	}

	if len(paths) == 0 {
		return "", fmt.Errorf("no paths provided")
	}

	for _, path := range paths {
		for _, fe := range ExecutableFileExtensions {
			fullpath := filepath.Join(path, plugin) + fe
			if fi, err := os.Stat(fullpath); err == nil && fi.Mode().IsRegular() {
				return fullpath, nil
			}
		}
	}

	return "", fmt.Errorf("failed to find plugin %q in path %s", plugin, paths)
}
