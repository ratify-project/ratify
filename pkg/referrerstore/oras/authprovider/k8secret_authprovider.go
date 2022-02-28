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
	"os"

	"github.com/docker/cli/cli/config"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type k8SecretProviderFactory struct{}
type k8SecretAuthProvider struct {
	secrets []*core.Secret
}

type secretConfig struct {
	SecretName string `json:"secretName"`
	Namespace  string `json:"namespace,omitempty"`
}

type k8SecretAuthProviderConf struct {
	Name               string         `json:"name"`
	ServiceAccountName string         `json:"serviceAccountName,omitempty"`
	Secrets            []secretConfig `json:"secrets,omitempty"`
}

const defaultName = "default"
const ratifyNamespaceEnv = "RATIFY_NAMESPACE"

// init calls Register for our k8s-secrets provider
func init() {
	Register("k8s-secrets", &k8SecretProviderFactory{})
}

// Create returns a k8AuthProvider instance after parsing auth config and resolving
// named K8 secrets
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

	if conf.ServiceAccountName == "" {
		conf.ServiceAccountName = defaultName
	}

	// get name of namespace ratify is running in
	ratifyNamespace := os.Getenv(ratifyNamespaceEnv)
	if ratifyNamespace == "" {
		return nil, fmt.Errorf("environment variable %s not set", ratifyNamespaceEnv)
	}

	var k8secrets []*core.Secret
	ctx := context.Background()

	// iterate through config secrets, resolve each secret, and store in map
	for _, k8secret := range conf.Secrets {
		// default value of secret is assumed to be ratify namespace
		if k8secret.Namespace == "" {
			k8secret.Namespace = ratifyNamespace
		}

		secret, err := clusterClientSet.CoreV1().Secrets(k8secret.Namespace).Get(ctx, k8secret.SecretName, meta.GetOptions{})
		if err != nil {
			return nil, err
		}

		// only dockercfg or docker config json secret type allowed
		if secret.Type == core.SecretTypeDockercfg || secret.Type == core.SecretTypeDockerConfigJson {
			k8secrets = append(k8secrets, secret)
		} else {
			return nil, fmt.Errorf("secret with unsupported type %s provided in config", secret.Type)
		}
	}

	// get the the service account for ratify
	serviceAccount, err := clusterClientSet.CoreV1().ServiceAccounts(ratifyNamespace).Get(ctx, conf.ServiceAccountName, meta.GetOptions{})
	if err != nil {
		return nil, err
	}

	// extract the imagePullSecrets linked to service account
	for _, imagePullSecret := range serviceAccount.ImagePullSecrets {
		secret, err := clusterClientSet.CoreV1().Secrets(ratifyNamespace).Get(ctx, imagePullSecret.Name, meta.GetOptions{})
		if err != nil {
			return nil, err
		}

		// only dockercfg or docker config json secret type allowed
		if secret.Type == core.SecretTypeDockercfg || secret.Type == core.SecretTypeDockerConfigJson {
			k8secrets = append(k8secrets, secret)
		}
	}

	return &k8SecretAuthProvider{
		secrets: k8secrets,
	}, nil
}

// Enabled checks if secrets list is not nil or empty
func (d *k8SecretAuthProvider) Enabled() bool {
	if d.secrets == nil || len(d.secrets) <= 0 {
		return false
	}

	return true
}

// Provide finds secret corresponding to artifact's registry host name, extracts
// the authentication credentials from k8 secret, and returns AuthConfig
func (d *k8SecretAuthProvider) Provide(artifact string) (AuthConfig, error) {
	if !d.Enabled() {
		return AuthConfig{}, fmt.Errorf("k8 secret provider not properly enabled")
	}

	hostName, err := getRegistryHostName(artifact)
	if err != nil {
		return AuthConfig{}, err
	}

	for _, secretLoaded := range d.secrets {
		if secretLoaded.Type == core.SecretTypeDockercfg {
			// if secret is a legacy docker config type
			dockercfg, exists := secretLoaded.Data[core.DockerConfigKey]
			if !exists {
				return AuthConfig{}, fmt.Errorf("could not extract auth configs from .dockercfg")
			}

			configFile, err := config.LegacyLoadFromReader(bytes.NewReader(dockercfg))
			if err != nil {
				return AuthConfig{}, err
			}

			authConfig, exist := configFile.AuthConfigs[hostName]
			if exist {
				return AuthConfig{
					Username: authConfig.Username,
					Password: authConfig.Password,
					Provider: d,
				}, nil
			}
		} else if secretLoaded.Type == core.SecretTypeDockerConfigJson {
			// if secret is a docker config json type
			dockerconfig, exists := secretLoaded.Data[core.DockerConfigJsonKey]
			if !exists {
				return AuthConfig{}, fmt.Errorf("could not extract auth configs from .docker/config.json")
			}

			configFile, err := config.LoadFromReader(bytes.NewReader(dockerconfig))
			if err != nil {
				return AuthConfig{}, err
			}

			authConfig, exist := configFile.AuthConfigs[hostName]
			if exist {
				return AuthConfig{
					Username: authConfig.Username,
					Password: authConfig.Password,
					Provider: d,
				}, nil
			}
		}
	}

	return AuthConfig{}, fmt.Errorf("could not find credentials for %s", artifact)
}
