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
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/configfile"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type k8SecretProviderFactory struct{}
type k8SecretAuthProvider struct {
	ratifyNamespace  string
	config           k8SecretAuthProviderConf
	clusterClientSet *kubernetes.Clientset
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
const secretTimeout = time.Hour * 12

var ErrorNoMatchingCredential = errors.New("no matching credential found for k8 secret")

// init calls Register for our k8Secrets provider
func init() {
	Register("k8Secrets", &k8SecretProviderFactory{})
}

// Create returns a k8AuthProvider instance after parsing auth config and resolving
// named K8 secrets
func (s *k8SecretProviderFactory) Create(authProviderConfig AuthProviderConfig) (AuthProvider, error) {

	conf := k8SecretAuthProviderConf{}
	authProviderConfigBytes, err := json.Marshal(authProviderConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal authentication provider config: %v", err)
	}

	if err := json.Unmarshal(authProviderConfigBytes, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse authentication provider configuration: %v", err)
	}

	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to generate cluster configuration: %v", err)
	}

	clientSet, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client set from config: %v", err)
	}

	if conf.ServiceAccountName == "" {
		conf.ServiceAccountName = defaultName
	}

	// get name of namespace ratify is running in
	namespace := os.Getenv(ratifyNamespaceEnv)
	if namespace == "" {
		return nil, fmt.Errorf("environment variable %s not set", ratifyNamespaceEnv)
	}

	return &k8SecretAuthProvider{
		ratifyNamespace:  namespace,
		config:           conf,
		clusterClientSet: clientSet,
	}, nil
}

// Enabled checks if ratify namespace, config, or cluster client set is nil
func (d *k8SecretAuthProvider) Enabled(ctx context.Context) bool {
	if d.ratifyNamespace == "" || d.clusterClientSet == nil {
		return false
	}

	return true
}

// Provide finds secret corresponding to artifact's registry host name, extracts
// the authentication credentials from k8 secret, and returns AuthConfig
func (d *k8SecretAuthProvider) Provide(ctx context.Context, artifact string) (AuthConfig, error) {
	if !d.Enabled(ctx) {
		return AuthConfig{}, fmt.Errorf("k8 auth provider not properly enabled")
	}

	hostName, err := GetRegistryHostName(artifact)
	if err != nil {
		return AuthConfig{}, err
	}

	// iterate through config secrets, resolve each secret, and store in map
	for _, k8secret := range d.config.Secrets {
		// default value of secret is assumed to be ratify namespace
		if k8secret.Namespace == "" {
			k8secret.Namespace = d.ratifyNamespace
		}

		secret, err := d.clusterClientSet.CoreV1().Secrets(k8secret.Namespace).Get(ctx, k8secret.SecretName, meta.GetOptions{})
		if err != nil {
			return AuthConfig{}, fmt.Errorf("failed to pull secret %s from cluster: %v", k8secret.SecretName, err)
		}

		// only dockercfg or docker config json secret type allowed
		if secret.Type == core.SecretTypeDockercfg || secret.Type == core.SecretTypeDockerConfigJson {
			authConfig, err := d.resolveCredentialFromSecret(hostName, secret)
			if err != nil && err != ErrorNoMatchingCredential {
				return AuthConfig{}, err
			}
			// if a resolved AuthConfig is returned
			if err == nil {
				return authConfig, nil
			}
		} else {
			return AuthConfig{}, fmt.Errorf("secret with unsupported type %s provided in config", secret.Type)
		}
	}

	// get the the service account for ratify
	serviceAccount, err := d.clusterClientSet.CoreV1().ServiceAccounts(d.ratifyNamespace).Get(ctx, d.config.ServiceAccountName, meta.GetOptions{})
	if err != nil {
		return AuthConfig{}, err
	}

	// extract the imagePullSecrets linked to service account
	for _, imagePullSecret := range serviceAccount.ImagePullSecrets {
		secret, err := d.clusterClientSet.CoreV1().Secrets(d.ratifyNamespace).Get(ctx, imagePullSecret.Name, meta.GetOptions{})
		if err != nil {
			return AuthConfig{}, err
		}

		// only dockercfg or docker config json secret type allowed
		if secret.Type == core.SecretTypeDockercfg || secret.Type == core.SecretTypeDockerConfigJson {
			authConfig, err := d.resolveCredentialFromSecret(hostName, secret)
			if err != nil && err != ErrorNoMatchingCredential {
				return AuthConfig{}, err
			}
			// if a resolved AuthConfig is returned
			if err == nil {
				return authConfig, nil
			}
		}
	}

	return AuthConfig{}, fmt.Errorf("could not find credentials for %s", artifact)
}

func (d *k8SecretAuthProvider) resolveCredentialFromSecret(hostName string, secret *core.Secret) (AuthConfig, error) {
	var secretDataKey string
	var configFileFn func(io.Reader) (*configfile.ConfigFile, error)
	if secret.Type == core.SecretTypeDockercfg {
		// if secret is a legacy docker config type
		secretDataKey = core.DockerConfigKey
		configFileFn = config.LegacyLoadFromReader
	} else {
		// if secret is a docker config json type
		secretDataKey = core.DockerConfigJsonKey
		configFileFn = config.LoadFromReader
	}

	dockercfg, exists := secret.Data[secretDataKey]
	if !exists {
		return AuthConfig{}, fmt.Errorf("could not extract auth configs from docker config")
	}

	configFile, err := configFileFn(bytes.NewReader(dockercfg))
	if err != nil {
		return AuthConfig{}, err
	}

	authConfig, exist := configFile.AuthConfigs[hostName]
	if !exist {
		return AuthConfig{}, ErrorNoMatchingCredential
	}

	return AuthConfig{
		Username:  authConfig.Username,
		Password:  authConfig.Password,
		Provider:  d,
		ExpiresOn: time.Now().Add(secretTimeout),
	}, nil
}
