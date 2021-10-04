package factory

import (
	"errors"
	"fmt"

	"github.com/deislabs/hora/pkg/verifier"
	"github.com/deislabs/hora/pkg/verifier/config"
	"github.com/deislabs/hora/pkg/verifier/plugin"
	"github.com/deislabs/hora/pkg/verifier/types"
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

// TODO pointer to avoid copy
func CreateVerifiersFromConfig(verifiersConfig config.VerifiersConfig, defaultPluginPath string) ([]verifier.ReferenceVerifier, error) {
	if verifiersConfig.Version == "" {
		verifiersConfig.Version = types.SpecVersion
	}

	err := validateVerifiersConfig(&verifiersConfig)
	if err != nil {
		return nil, err
	}

	if len(verifiersConfig.Verifiers) == 0 {
		return nil, errors.New("verifiers config should have atleast one verifer")
	}

	var verifiers []verifier.ReferenceVerifier

	if len(verifiersConfig.PluginBinDirs) == 0 {
		verifiersConfig.PluginBinDirs = []string{defaultPluginPath}
	}
	for _, verifierConfig := range verifiersConfig.Verifiers {
		verifierName, ok := verifierConfig[types.Name]
		if !ok {
			return nil, fmt.Errorf("failed to find verifier name in the verifier config with key %s", "name")
		}

		verifierFactory, ok := builtInVerifiers[fmt.Sprintf("%s", verifierName)]
		if ok {
			verifier, err := verifierFactory.Create(verifiersConfig.Version, verifierConfig)

			if err != nil {
				return nil, err
			}

			verifiers = append(verifiers, verifier)
		} else {
			verifier, err := plugin.NewVerifier(verifiersConfig.Version, verifierConfig, verifiersConfig.PluginBinDirs)

			if err != nil {
				return nil, err
			}

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
