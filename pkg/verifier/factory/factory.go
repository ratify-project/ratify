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

package factory

import (
	"errors"
	"fmt"
	"os"
	"strings"

	pluginCommon "github.com/deislabs/ratify/pkg/common/plugin"
	"github.com/deislabs/ratify/pkg/featureflag"
	"github.com/deislabs/ratify/pkg/verifier"
	"github.com/deislabs/ratify/pkg/verifier/config"
	"github.com/deislabs/ratify/pkg/verifier/plugin"
	"github.com/deislabs/ratify/pkg/verifier/types"
	"github.com/sirupsen/logrus"
)

var builtInVerifiers = make(map[string]VerifierFactory)

type VerifierFactory interface {
	Create(version string, verifierConfig config.VerifierConfig) (verifier.ReferenceVerifier, error)
}

func Register(name string, factory VerifierFactory) {
	if factory == nil {
		panic("Verifier factor cannot be nil")
	}
	_, registered := builtInVerifiers[name]
	if registered {
		panic(fmt.Sprintf("verifier factory named %s already registered", name))
	}

	builtInVerifiers[name] = factory
}

// returns a single verifier from a verifierConfig
func CreateVerifierFromConfig(verifierConfig config.VerifierConfig, configVersion string, pluginBinDir []string) (verifier.ReferenceVerifier, error) {

	verifierName, ok := verifierConfig[types.Name]
	if !ok {
		return nil, fmt.Errorf("failed to find verifier name in the verifier config with key %s", "name")
	}

	verifierNameStr := fmt.Sprintf("%s", verifierName)
	if strings.ContainsRune(verifierNameStr, os.PathSeparator) {
		return nil, fmt.Errorf("invalid plugin name for a verifier: %s", verifierNameStr)
	}

	// if source is specified, download the plugin
	if source, ok := verifierConfig[types.Source]; ok && source != "" {
		if featureflag.DynamicPlugins.Enabled {
			sourceStr := fmt.Sprintf("%s", source)
			err := pluginCommon.DownloadPlugin(verifierNameStr, sourceStr, pluginBinDir[0])
			if err != nil {
				return nil, fmt.Errorf("failed to download plugin: %v", err)
			}
			logrus.Infof("downloaded verifier plugin %s from %s to %s", verifierNameStr, sourceStr, pluginBinDir[0])
		} else {
			logrus.Warnf("%s was specified for verifier plugin %s, but dynamic plugins are currently disabled", types.Source, verifierNameStr)
		}
	}

	verifierFactory, ok := builtInVerifiers[verifierNameStr]
	if ok {
		return verifierFactory.Create(configVersion, verifierConfig)
	} else {
		return plugin.NewVerifier(configVersion, verifierConfig, pluginBinDir)
	}
}

// TODO pointer to avoid copy
// returns an array of verifiers from VerifiersConfig
func CreateVerifiersFromConfig(verifiersConfig config.VerifiersConfig, defaultPluginPath string) ([]verifier.ReferenceVerifier, error) {
	if verifiersConfig.Version == "" {
		verifiersConfig.Version = types.SpecVersion
	}

	err := validateVerifiersConfig(&verifiersConfig)
	if err != nil {
		return nil, err
	}

	if len(verifiersConfig.Verifiers) == 0 {
		return nil, errors.New("verifiers config should have at least one verifier")
	}

	var verifiers []verifier.ReferenceVerifier

	if len(verifiersConfig.PluginBinDirs) == 0 {
		verifiersConfig.PluginBinDirs = []string{defaultPluginPath}
		logrus.Info("defaultPluginPath set to " + defaultPluginPath)
	}

	// TODO: do we need to append defaultPlugin path?
	for _, verifierConfig := range verifiersConfig.Verifiers {
		verifier, err := CreateVerifierFromConfig(verifierConfig, verifiersConfig.Version, verifiersConfig.PluginBinDirs)
		if err != nil {
			return nil, err
		} else {
			verifiers = append(verifiers, verifier)
		}
	}

	return verifiers, nil
}

func validateVerifiersConfig(verifiersConfig *config.VerifiersConfig) error {
	// TODO check for existence of plugin dirs
	// TODO check if version is supported
	return nil

}
