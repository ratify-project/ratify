//go:build !windows

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

package config

import (
	"os"
	"os/user"
	"testing"
)

func TestGet(t *testing.T) {
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	t.Run("HOME environment variable set", func(t *testing.T) {
		expectedHome := "/tmp/testhome"
		os.Setenv("HOME", expectedHome)
		if home := get(); home != expectedHome {
			t.Errorf("expected home %q, got %q", expectedHome, home)
		}
	})

	t.Run("HOME environment variable unset, fallback to user.Current()", func(t *testing.T) {
		os.Unsetenv("HOME")
		currentUser, err := user.Current()
		if err != nil {
			t.Fatalf("failed to get current user: %v", err)
		}
		if home := get(); home != currentUser.HomeDir {
			t.Errorf("expected home %q, got %q", currentUser.HomeDir, home)
		}
	})
}
