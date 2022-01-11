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

package authprovider

import (
	"github.com/containerd/containerd/reference"
	"github.com/docker/cli/cli/config"
)

// This config that represents the credentials that should be used
// when pulling artifacts from specific repositories.
type AuthConfig struct {
	Username string
	Password string
	Email    string       // NOT SURE IF NECESSARY
	Provider AuthProvider // NOT SURE IF NECESSARY
}

type AuthProvider interface {
	// Enabled returns true if the config provider is enabled.
	// Implementations can be blocking - e.g. MI config file not found
	Enabled() bool
	// Provide returns AuthConfig for registry.
	// Implementations can be blocking - e.g. MI config not found.
	// The artifact is passed in as context in the event that the
	// implementation depends on information in the artifact name to return
	// credentials; implementations are safe to ignore the artifact.
	Provide(artifact string) AuthConfig
}

type defaultProviderFactory struct{}
type defaultAuthProvider struct{}

// init calls Register for our default provider, which simply reads the .dockercfg file.
func init() {
	Register("default", &defaultProviderFactory{})
}

// Create defaultAuthProvider
func (s *defaultProviderFactory) Create(authProviderConfig AuthProviderConfig) (AuthProvider, error) {
	return &defaultAuthProvider{}, nil
}

// Enabled implements AuthProvider; Always returns true
func (d *defaultAuthProvider) Enabled() bool {
	return true
}

// Provide implements AuthProvider; reads docker config file and returns corresponding credentials from file if exists
func (d *defaultAuthProvider) Provide(artifact string) AuthConfig {
	cfg, err := config.Load(config.Dir())
	if err != nil {
		return AuthConfig{}
	}
	parsedSpec, err := reference.Parse(artifact)
	if err != nil {
		return AuthConfig{}
	}
	artifactHostName := parsedSpec.Hostname()

	dockerAuthConfig := cfg.AuthConfigs[artifactHostName]
	authConfig := AuthConfig{
		Username: dockerAuthConfig.Username,
		Password: dockerAuthConfig.Password,
		Provider: d,
	}
	return authConfig
}
