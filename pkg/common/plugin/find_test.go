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
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindInPaths_FindsPluginInPaths(t *testing.T) {
	// Create a temp directory
	tempDir, err := os.MkdirTemp("", "plugin-find")
	if err != nil {
		t.Fatalf("temp directory creation failed %v", err)
	}

	defer func() {
		os.RemoveAll(tempDir)
	}()

	plugin1, err := os.CreateTemp(tempDir, "test-plugin1")

	if err != nil {
		t.Fatalf("test plugin1 creation failed %v", err)
	}

	pluginWithOSSpecificExtName := "test-plugin2" + executableFileExtensions[0]
	plugin2, err := os.Create(filepath.Join(tempDir, pluginWithOSSpecificExtName))

	if err != nil {
		t.Fatalf("test plugin2 creation failed %v", err)
	}

	pluginDir, plugin1Name := filepath.Split(plugin1.Name())

	plugin1Path, err := FindInPaths(plugin1Name, []string{pluginDir})
	if err != nil {
		t.Fatalf("find plugin failed %v", err)
	}

	expected := filepath.Join(pluginDir, plugin1Name)
	if plugin1Path != expected {
		t.Fatalf("plugin path found mismatches expected: %v got: %v", expected, plugin1Path)
	}

	_, plugin2Name := filepath.Split(plugin2.Name())
	plugin2NameWithoutExt := strings.Split(plugin2Name, ".")[0]
	plugin2Path, err := FindInPaths(plugin2NameWithoutExt, []string{pluginDir})
	if err != nil {
		t.Fatalf("find plugin failed %v", err)
	}

	expected = filepath.Join(pluginDir, plugin2Name)
	if plugin2Path != expected {
		t.Fatalf("plugin path found mismatches expected: %v got: %v", expected, plugin2Path)
	}
}

func TestFindInPaths_ExpectedErrors(t *testing.T) {
	// Create a temp directory
	tempDir, err := os.MkdirTemp("", "plugin-empty")
	if err != nil {
		t.Fatalf("temp directory creation failed %v", err)
	}

	defer func() {
		os.RemoveAll(tempDir)
	}()

	_, err = FindInPaths("", []string{})
	if err == nil || !strings.Contains(err.Error(), "plugin name is required") {
		t.Fatalf("expected error 'plugin name is required' not returned %v", err)
	}

	_, err = FindInPaths("test-plugin", []string{})
	if err == nil || !strings.Contains(err.Error(), "no paths provided to find a plugin") {
		t.Fatalf("expected error 'no paths provided to find a plugin' not returned %v", err)
	}

	_, err = FindInPaths("test-plugin", []string{tempDir})
	if err == nil || !strings.Contains(err.Error(), "failed to find plugin") {
		t.Fatalf("expected error 'failed to find plugin' not returned %v", err)
	}
}
