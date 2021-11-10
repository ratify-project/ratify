package plugin

import (
	"os"
	"testing"
)

func TestAsEnviron_ReturnsExpected(t *testing.T) {
	existingEnv := os.Environ()
	countOfExistingEnv := len(existingEnv)

	args := ReferrerStorePluginArgs{
		Command:          "testCommand",
		Version:          "1.0.0",
		SubjectReference: "testref",
		PluginArgs: [][2]string{
			{"testkey1", "testvalue1"},
		},
	}

	storePluginArgs := args.AsEnviron()
	if countOfExistingEnv+4 != len(storePluginArgs) {
		t.Fatalf("mismatch of the plugin env")
	}

	hasEnv := func(env string) bool {
		for _, e := range storePluginArgs {
			if e == env {
				return true
			}
		}

		return false
	}

	if !hasEnv("RATIFY_STORE_COMMAND=testCommand") {
		t.Fatalf("missing command env")
	}

	if !hasEnv("RATIFY_STORE_SUBJECT=testref") {
		t.Fatalf("missing subject env")
	}

	if !hasEnv("RATIFY_STORE_ARGS=testkey1=testvalue1") {
		t.Fatalf("missing args env")
	}

	if !hasEnv("RATIFY_STORE_VERSION=1.0.0") {
		t.Fatalf("missing version env")
	}
}
