package config

import "github.com/deislabs/ratify/pkg/policyprovider/types"

type PoliciesConfig struct {
	Version                      string                                    `json:"version,omitempty"`
	ArtifactVerificationPolicies map[string]types.ArtifactTypeVerifyPolicy `json:"artifactVerificationPolicies,omitempty"`
}
