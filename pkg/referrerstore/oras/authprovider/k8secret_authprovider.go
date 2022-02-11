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
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/docker/cli/cli/config"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type k8SecretProviderFactory struct{}
type k8SecretAuthProvider struct {
	secrets map[string]loadedSecret
}

type secretConfig struct {
	RegistryHost string `json:"registryHost"`
	SecretName   string `json:"secretName"`
	Namespace    string `json:"namespace,omitempty"`
}

type k8SecretAuthProviderConf struct {
	Name    string         `json:"name"`
	Secrets []secretConfig `json:"secrets,omitempty"`
}

type loadedSecret struct {
	RegistryHost string
	Secret       *core.Secret
}

// init calls Register for our k8s-secrets provider
func init() {
	Register("k8s-secrets", &k8SecretProviderFactory{})
}

// Create returns an empty defaultAuthProvider instance if the AuthProviderConfig is nil.
// Otherwise it returns the defaultAuthProvider with configPath set
func (s *k8SecretProviderFactory) Create(authProviderConfig AuthProviderConfig) (AuthProvider, error) {

	conf := k8SecretAuthProviderConf{}
	authProviderConfigBytes, err := json.Marshal(authProviderConfig)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(authProviderConfigBytes, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse auth provider configuration: %v", err)
	}

	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clusterClientSet, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return nil, err
	}

	var extractedSecrets = make(map[string]loadedSecret)
	// iterate through configuration secrets,resolve each secret, and store in map
	for _, secretConf := range conf.Secrets {
		if secretConf.Namespace == "" {
			secretConf.Namespace = "default"
		}

		// each registry host specified must be unique
		if _, ok := extractedSecrets[secretConf.RegistryHost]; ok {
			return nil, fmt.Errorf("registry host %s already has configured secret", secretConf.RegistryHost)
		}

		secret, err := clusterClientSet.CoreV1().Secrets(secretConf.Namespace).Get(context.Background(), secretConf.SecretName, meta.GetOptions{})
		if err != nil {
			return nil, err
		}
		extractedSecrets[secretConf.RegistryHost] = loadedSecret{
			RegistryHost: secretConf.RegistryHost,
			Secret:       secret,
		}
	}

	return &k8SecretAuthProvider{
		secrets: extractedSecrets,
	}, nil
}

// Enabled always returns true for defaultAuthProvider
func (d *k8SecretAuthProvider) Enabled() bool {
	if d.secrets == nil || len(d.secrets) <= 0 {
		return false
	}

	return true
}

// Provide finds secret corresponding to artifact's registryhost,
// extracts the authentication credentials from k8 secret, and
// returns AuthConfig
func (d *k8SecretAuthProvider) Provide(artifact string) (AuthConfig, error) {
	hostName, err := getRegistryHostName(artifact)
	if err != nil {
		return AuthConfig{}, err
	}

	secretLoaded, exists := d.secrets[hostName]
	if !exists {
		return AuthConfig{}, fmt.Errorf("could not find secret corresponding for artifact: %s", artifact)
	}

	if secretLoaded.Secret.Type == core.SecretTypeBasicAuth {
		// if secret is of type basic-auth
		return AuthConfig{
			Username: string(secretLoaded.Secret.Data[core.BasicAuthUsernameKey]),
			Password: string(secretLoaded.Secret.Data[core.BasicAuthPasswordKey]),
			Provider: d,
		}, nil
	} else if secretLoaded.Secret.Type == core.SecretTypeDockercfg {
		// if secret is a legacy docker config type
		dockercfg, exists := secretLoaded.Secret.Data[core.DockerConfigKey]
		if !exists {
			return AuthConfig{}, fmt.Errorf("could not extract auth configs from .dockercfg")
		}

		configFile, err := config.LegacyLoadFromReader(bytes.NewReader(dockercfg))
		if err != nil {
			return AuthConfig{}, err
		}

		authConfig, exist := configFile.AuthConfigs[hostName]
		if !exist {
			return AuthConfig{}, fmt.Errorf("could not find credentials for %s in .dockercfg", hostName)
		}

		return AuthConfig{
			Username: authConfig.Username,
			Password: authConfig.Password,
			Provider: d,
		}, nil
	} else if secretLoaded.Secret.Type == core.SecretTypeDockerConfigJson {
		// if secret is a docker config json type
		dockerconfig, exists := secretLoaded.Secret.Data[core.DockerConfigJsonKey]
		if !exists {
			return AuthConfig{}, fmt.Errorf("could not extract auth configs from .docker/config.json")
		}

		configFile, err := config.LoadFromReader(bytes.NewReader(dockerconfig))
		if err != nil {
			return AuthConfig{}, err
		}

		authConfig, exist := configFile.AuthConfigs[hostName]
		if !exist {
			return AuthConfig{}, fmt.Errorf("could not find credentials for %s in config.json", hostName)
		}

		return AuthConfig{
			Username: authConfig.Username,
			Password: authConfig.Password,
			Provider: d,
		}, nil
	}

	return AuthConfig{}, nil
}
