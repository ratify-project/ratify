package plugin

import (
	"os"
	"testing"
)

func TestAsEnviron_ReturnsExpected(t *testing.T) {
	existingEnv := os.Environ()
	countOfExistingEnv := len(existingEnv)

	args := VerifierPluginArgs{
		Command:          "testCommand",
		Version:          "1.0.0",
		SubjectReference: "testref",
	}

	verifierPluginArgs := args.AsEnviron()
	if countOfExistingEnv+3 != len(verifierPluginArgs) {
		t.Fatalf("mismatch of the plugin env")
	}

	hasEnv := func(env string) bool {
		for _, e := range verifierPluginArgs {
			if e == env {
				return true
			}
		}

		return false
	}

	if !hasEnv("RATIFY_VERIFIER_COMMAND=testCommand") {
		t.Fatalf("missing command env")
	}

	if !hasEnv("RATIFY_VERIFIER_SUBJECT=testref") {
		t.Fatalf("missing subject env")
	}

	if !hasEnv("RATIFY_VERIFIER_VERSION=1.0.0") {
		t.Fatalf("missing version env")
	}
}
