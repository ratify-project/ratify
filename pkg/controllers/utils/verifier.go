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

package utils

import (
	"encoding/json"

	configv1beta1 "github.com/ratify-project/ratify/api/v1beta1"
	vc "github.com/ratify-project/ratify/pkg/verifier/config"
	vf "github.com/ratify-project/ratify/pkg/verifier/factory"
	"github.com/ratify-project/ratify/pkg/verifier/types"

	"github.com/ratify-project/ratify/config"
	"github.com/ratify-project/ratify/pkg/controllers"
	"github.com/sirupsen/logrus"
)

// UpsertVerifier creates and adds a new verifier based on the provided configuration.
func UpsertVerifier(version, address, namespace, objectName string, verifierConfig vc.VerifierConfig) error {
	if len(version) == 0 {
		version = config.GetDefaultPluginVersion()
		logrus.Infof("Version was empty, setting to default version: %v", version)
	}

	if address == "" {
		address = config.GetDefaultPluginPath()
		logrus.Infof("Address was empty, setting to default path: %v", address)
	}

	referenceVerifier, err := vf.CreateVerifierFromConfig(verifierConfig, version, []string{address}, namespace)
	if err != nil || referenceVerifier == nil {
		logrus.Error(err, " unable to create verifier from verifier config")
		return err
	}

	controllers.NamespacedVerifiers.AddVerifier(namespace, objectName, referenceVerifier)
	logrus.Infof("verifier '%v' added to verifier map in namespace: %s", referenceVerifier.Name(), namespace)

	return nil
}

// SpecToVerifierConfig returns a VerifierConfig from VerifierSpec
func SpecToVerifierConfig(raw []byte, verifierName, verifierType, artifactTypes string, source *configv1beta1.PluginSource) (vc.VerifierConfig, error) {
	verifierConfig := vc.VerifierConfig{}

	if string(raw) != "" {
		if err := json.Unmarshal(raw, &verifierConfig); err != nil {
			logrus.Error(err, "unable to decode verifier parameters", "Parameters.Raw", raw)
			return vc.VerifierConfig{}, err
		}
	}
	verifierConfig[types.Name] = verifierName
	verifierConfig[types.Type] = verifierType
	verifierConfig[types.ArtifactTypes] = artifactTypes
	if source != nil {
		verifierConfig[types.Source] = source
	}

	return verifierConfig, nil
}
