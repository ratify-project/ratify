package featureflag

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Feature flags are used to enable/disable experimental features.
// They are activated via environment variables, starting with "RATIFY_", ex: RATIFY_DYNAMIC_PLUGINS=1
// Remember to capture changes in the usage guide and release notes.
var (
	DynamicPlugins = new("DYNAMIC_PLUGINS", false)
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

func new(name string, defaultValue bool) *FeatureFlag {
	flag := &FeatureFlag{
		Name: name,
		Enabled: defaultValue,
	}
	flags[name] = flag
	return flag
}
