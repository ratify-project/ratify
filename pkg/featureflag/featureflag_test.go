package featureflag

import (
	"os"
	"testing"
)

func TestFeatureFlag_UsesDefaultValues(t *testing.T) {
	flag1 := new("TEST_FLAG1", false)
	if flag1.Enabled {
		t.Fatal("expected flag1 to be disabled")
	}

	flag2 := new("TEST_FLAG2", true)
	if !flag2.Enabled {
		t.Fatal("expected flag2 to be enabled")
	}
}

func TestInitFeatureFlagsFromEnv_SetsValues(t *testing.T) {
	flag1 := new("TEST_FLAG1", false)
	flag2 := new("TEST_FLAG2", false)
	flag3 := new("TEST_FLAG3", true)
	flag4 := new("TEST_FLAG4", true)

	// override flag defaults
	os.Setenv("RATIFY_TEST_FLAG2", "1")
	os.Setenv("RATIFY_TEST_FLAG4", "0")

	InitFeatureFlagsFromEnv()

	if flag1.Enabled {
		t.Fatal("expected flag1 to be disabled")
	}

	if !flag2.Enabled {
		t.Fatal("expected flag2 to be enabled")
	}

	if !flag3.Enabled {
		t.Fatal("expected flag3 to be enabled")
	}

	if flag4.Enabled {
		t.Fatal("expected flag4 to be disabled")
	}
}
