package factory

import (
	"errors"

	"github.com/notaryproject/hora/pkg/verifier"
	"github.com/notaryproject/hora/pkg/verifier/config"
	"github.com/notaryproject/hora/pkg/verifier/plugin"
	"github.com/notaryproject/hora/pkg/verifier/types"
)

// TODO pointer to avoid copy
func CreateVerifiersFromConfig(verifiersConfig config.VerifiersConfig, defaultPluginPath string) ([]verifier.ReferenceVerifier, error) {
	if verifiersConfig.Version == "" {
		verifiersConfig.Version = types.SpecVersion
	}

	err := validateVerifiersConfig(&verifiersConfig)
	if err != nil {
		return nil, err
	}

	/*if len(verifiersConfig.PluginBinDirs) == 0 {
		return nil, errors.New("atleast one plugin bin directory is required")
	}*/

	if len(verifiersConfig.Verifiers) == 0 {
		return nil, errors.New("verifiers config should have atleast one verifer")
	}

	var verifiers []verifier.ReferenceVerifier

	if len(verifiersConfig.PluginBinDirs) == 0 {
		verifiersConfig.PluginBinDirs = []string{defaultPluginPath}
	}
	for _, verifierConfig := range verifiersConfig.Verifiers {

		verifer, err := plugin.NewVerifier(verifiersConfig.Version, verifierConfig, verifiersConfig.PluginBinDirs)

		if err != nil {
			return nil, err
		}

		verifiers = append(verifiers, verifer)
	}

	return verifiers, nil
}

func validateVerifiersConfig(verifiersConfig *config.VerifiersConfig) error {
	// TODO check for existence of plugin dirs
	// TODO check if version is supported
	return nil

}
