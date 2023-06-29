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

package featureflag

import (
	"os"
	"testing"
)

func TestFeatureFlag_UsesDefaultValues(t *testing.T) {
	flag1 := newFeatureFlag("TEST_FLAG1", false)
	if flag1.Enabled {
		t.Fatal("expected flag1 to be disabled")
	}

	flag2 := newFeatureFlag("TEST_FLAG2", true)
	if !flag2.Enabled {
		t.Fatal("expected flag2 to be enabled")
	}
}

func TestInitFeatureFlagsFromEnv_SetsValues(t *testing.T) {
	flag1 := newFeatureFlag("TEST_FLAG1", false)
	flag2 := newFeatureFlag("TEST_FLAG2", false)
	flag3 := newFeatureFlag("TEST_FLAG3", true)
	flag4 := newFeatureFlag("TEST_FLAG4", true)

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
