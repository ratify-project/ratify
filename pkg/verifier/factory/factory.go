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
	"fmt"
	"os"
	"path"
	"strings"

	re "github.com/ratify-project/ratify/errors"
	pluginCommon "github.com/ratify-project/ratify/pkg/common/plugin"
	"github.com/ratify-project/ratify/pkg/featureflag"
	"github.com/ratify-project/ratify/pkg/verifier"
	"github.com/ratify-project/ratify/pkg/verifier/config"
	"github.com/ratify-project/ratify/pkg/verifier/plugin"
	"github.com/ratify-project/ratify/pkg/verifier/types"
	"github.com/sirupsen/logrus"
)

var builtInVerifiers = make(map[string]VerifierFactory)

type VerifierFactory interface {
	Create(version string, verifierConfig config.VerifierConfig, pluginDirectory string, namespace string) (verifier.ReferenceVerifier, error)
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
// namespace is only applicable in K8s environment, namespace is appended to the certstore of the truststore so it is uniquely identifiable in a cluster env
// the first element of pluginBinDir will be used as the plugin directory
func CreateVerifierFromConfig(verifierConfig config.VerifierConfig, configVersion string, pluginBinDir []string, namespace string) (verifier.ReferenceVerifier, error) {
	// in cli mode both `type` and `name`` are read from config, if `type` is not specified, `name` is used as `type`
	var verifierTypeStr string
	if value, ok := verifierConfig[types.Name]; ok {
		verifierTypeStr = value.(string)
	} else {
		return nil, re.ErrorCodeConfigInvalid.WithComponentType(re.Verifier).WithDetail(fmt.Sprintf("failed to find verifier name in the verifier config with key %s", "name"))
	}

	if value, ok := verifierConfig[types.Type]; ok {
		verifierTypeStr = value.(string)
	}

	if strings.ContainsRune(verifierTypeStr, os.PathSeparator) {
		return nil, re.ErrorCodeConfigInvalid.WithComponentType(re.Verifier).WithDetail(fmt.Sprintf("invalid plugin name for a verifier: %s", verifierTypeStr))
	}

	// if source is specified, download the plugin
	if source, ok := verifierConfig[types.Source]; ok {
		if featureflag.DynamicPlugins.Enabled {
			source, err := pluginCommon.ParsePluginSource(source)
			if err != nil {
				return nil, re.ErrorCodeConfigInvalid.NewError(re.Verifier, "", re.EmptyLink, err, "failed to parse plugin source", re.HideStackTrace)
			}

			targetPath := path.Join(pluginBinDir[0], verifierTypeStr)
			err = pluginCommon.DownloadPlugin(source, targetPath)
			if err != nil {
				return nil, re.ErrorCodeDownloadPluginFailure.NewError(re.Verifier, "", re.EmptyLink, err, "failed to download plugin", re.HideStackTrace)
			}
			logrus.Infof("downloaded verifier plugin %s from %s to %s", verifierTypeStr, source.Artifact, targetPath)
		} else {
			logrus.Warnf("%s was specified for verifier plugin type %s, but dynamic plugins are currently disabled", types.Source, verifierTypeStr)
		}
	}

	verifierFactory, ok := builtInVerifiers[verifierTypeStr]
	if ok {
		return verifierFactory.Create(configVersion, verifierConfig, pluginBinDir[0], namespace)
	}

	if _, err := pluginCommon.FindInPaths(verifierTypeStr, pluginBinDir); err != nil {
		return nil, re.ErrorCodePluginNotFound.NewError(re.Verifier, "", re.EmptyLink, err, "plugin not found", re.HideStackTrace)
	}

	return plugin.NewVerifier(configVersion, verifierConfig, pluginBinDir)
}

// TODO pointer to avoid copy
// returns an array of verifiers from VerifiersConfig
func CreateVerifiersFromConfig(verifiersConfig config.VerifiersConfig, defaultPluginPath string, namespace string) ([]verifier.ReferenceVerifier, error) {
	if verifiersConfig.Version == "" {
		verifiersConfig.Version = types.SpecVersion
	}

	err := validateVerifiersConfig(&verifiersConfig)
	if err != nil {
		return nil, re.ErrorCodeConfigInvalid.WithComponentType(re.Verifier).WithError(err)
	}

	if len(verifiersConfig.Verifiers) == 0 {
		return nil, re.ErrorCodeConfigInvalid.WithComponentType(re.Verifier).WithDetail("verifiers config should have at least one verifier")
	}

	verifiers := make([]verifier.ReferenceVerifier, 0)

	if len(verifiersConfig.PluginBinDirs) == 0 {
		verifiersConfig.PluginBinDirs = []string{defaultPluginPath}
		logrus.Info("defaultPluginPath set to " + defaultPluginPath)
	}

	// TODO: do we need to append defaultPlugin path?
	for _, verifierConfig := range verifiersConfig.Verifiers {
		verifier, err := CreateVerifierFromConfig(verifierConfig, verifiersConfig.Version, verifiersConfig.PluginBinDirs, namespace)
		if err != nil {
			return nil, re.ErrorCodePluginInitFailure.WithComponentType(re.Verifier).WithError(err)
		}
		verifiers = append(verifiers, verifier)
	}

	return verifiers, nil
}

func validateVerifiersConfig(_ *config.VerifiersConfig) error {
	// TODO check for existence of plugin dirs
	// TODO check if version is supported
	return nil
}
