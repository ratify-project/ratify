package config

import "github.com/deislabs/hora/pkg/policyprovider/types"

type PoliciesConfig struct {
	Version                      string                                    `yaml:"version,omitempty"`
	ArtifactVerificationPolicies map[string]types.ArtifactTypeVerifyPolicy `yaml:"artifactVerificationPolicies,omitempty"`
}
