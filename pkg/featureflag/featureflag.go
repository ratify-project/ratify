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

	"github.com/sirupsen/logrus"
)

// Feature flags are used to enable/disable experimental features.
// They are activated via environment variables, starting with "RATIFY_", ex: RATIFY_DYNAMIC_PLUGINS=1
// Remember to capture changes in the usage guide and release notes.
var (
	DynamicPlugins    = newFeatureFlag("DYNAMIC_PLUGINS", false)
	UseRegoPolicy     = newFeatureFlag("USE_REGO_POLICY", false)
	PassthroughMode   = newFeatureFlag("PASSTHROUGH_MODE", false)
	CertRotation      = newFeatureFlag("CERT_ROTATION", false)
	DaprCacheProvider = newFeatureFlag("DAPR_CACHE_PROVIDER", false)
)

var flags = make(map[string]*FeatureFlag)

func InitFeatureFlagsFromEnv() {
	for _, f := range flags {
		value, ok := os.LookupEnv("RATIFY_" + f.Name)
		if ok {
			f.Enabled = value == "1"

			if f.Enabled {
				logrus.Infof("Feature flag %s is enabled", f.Name)
			} else {
				logrus.Infof("Feature flag %s is disabled", f.Name)
			}
		}
	}
}

type FeatureFlag struct {
	Name    string
	Enabled bool
}

func newFeatureFlag(name string, defaultValue bool) *FeatureFlag {
	flag := &FeatureFlag{
		Name:    name,
		Enabled: defaultValue,
	}
	flags[name] = flag
	return flag
}
