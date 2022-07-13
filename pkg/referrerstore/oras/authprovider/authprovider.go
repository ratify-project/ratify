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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/configfile"
)

// This config represents the credentials that should be used
// when pulling artifacts from specific repositories.
type AuthConfig struct {
	Username  string
	Password  string
	Email     string
	Provider  AuthProvider
	ExpiresOn time.Time
}

type AuthProvider interface {
	// Enabled returns true if the config provider is properly enabled
	// It will verify necessary values provided in config file to
	// create the AuthProvider
	Enabled(ctx context.Context) bool
	// Provide returns AuthConfig for registry.
	Provide(ctx context.Context, artifact string) (AuthConfig, error)
}

type defaultProviderFactory struct{}
type defaultAuthProvider struct {
	configPath string
}

type defaultAuthProviderConf struct {
	Name       string `json:"name"`
	ConfigPath string `json:"configPath,omitempty"`
}

const DefaultAuthProviderName string = "dockerConfig"

// init calls Register for our default provider, which simply reads the .dockercfg file.
func init() {
	Register(DefaultAuthProviderName, &defaultProviderFactory{})
}

// Create returns an empty defaultAuthProvider instance if the AuthProviderConfig is nil.
// Otherwise it returns the defaultAuthProvider with configPath set
func (s *defaultProviderFactory) Create(authProviderConfig AuthProviderConfig) (AuthProvider, error) {
	if authProviderConfig == nil {
		return &defaultAuthProvider{configPath: ""}, nil
	}

	conf := defaultAuthProviderConf{}
	authProviderConfigBytes, err := json.Marshal(authProviderConfig)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(authProviderConfigBytes, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse auth provider configuration: %v", err)
	}

	return &defaultAuthProvider{
		configPath: conf.ConfigPath,
	}, nil
}

// Enabled always returns true for defaultAuthProvider
func (d *defaultAuthProvider) Enabled(ctx context.Context) bool {
	return true
}

// Provide reads docker config file and returns corresponding credentials from file if exists
func (d *defaultAuthProvider) Provide(ctx context.Context, artifact string) (AuthConfig, error) {
	// load docker config file at default path if config file path not specified
	var cfg *configfile.ConfigFile
	if d.configPath == "" {
		var err error
		cfg, err = config.Load(config.Dir())
		if err != nil {
			return AuthConfig{}, nil
		}
	} else {
		cfg = configfile.New(d.configPath)
		if _, err := os.Stat(d.configPath); err == nil {
			file, err := os.Open(d.configPath)
			if err != nil {
				return AuthConfig{}, err
			}
			defer file.Close()
			if err := cfg.LoadFromReader(file); err != nil {
				return AuthConfig{}, err
			}
		} else {
			return AuthConfig{}, err
		}
	}

	artifactHostName, err := GetRegistryHostName(artifact)
	if err != nil {
		return AuthConfig{}, err
	}

	dockerAuthConfig := cfg.AuthConfigs[artifactHostName]
	authConfig := AuthConfig{
		Username: dockerAuthConfig.Username,
		Password: dockerAuthConfig.Password,
		Provider: d,
	}

	return authConfig, nil
}

func GetRegistryHostName(artifact string) (string, error) {
	if strings.Contains(artifact, "://") {
		return "", errors.New("invalid artifact reference")
	}

	u, err := url.Parse("dummy://" + artifact)
	if err != nil {
		return "", err
	}

	if u.Scheme != "dummy" {
		return "", errors.New("invalid artifact reference: scheme missing")
	}

	if u.Host == "" {
		return "", errors.New("invalid artifact reference: host could not be extracted")
	}

	return u.Host, nil
}
