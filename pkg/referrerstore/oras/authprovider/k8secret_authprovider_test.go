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
	"fmt"
	"testing"

	core "k8s.io/api/core/v1"
)

// Checks K8 Basic-Auth Secret is properly extracted and credentials
// returned when Provide is called
func TestProvide_K8SecretBasicAuth_ReturnsExpected(t *testing.T) {
	var testSecret core.Secret
	testSecret.Data = make(map[string][]byte)
	testSecret.Type = core.SecretTypeBasicAuth
	testSecret.Data[core.BasicAuthUsernameKey] = []byte("test-username")
	testSecret.Data[core.BasicAuthPasswordKey] = []byte("test-password")

	var secretMap = make(map[string]*core.Secret)
	secretMap["test.ghcr.io"] = &testSecret
	var k8secretprovder k8SecretAuthProvider
	k8secretprovder.secrets = secretMap

	authConfig, err := k8secretprovder.Provide("test.ghcr.io/test-artifact:v1")
	if err != nil {
		t.Fatalf("provide failed to get credential with err %v", err)
	}

	if authConfig.Username != "test-username" || authConfig.Password != "test-password" {
		t.Fatalf("provide returned incorrect credentials (username: %s, password: %s)", authConfig.Username, authConfig.Password)
	}
}

// Checks K8 Docker Json Config Secret is properly extracted and
// credentials returned when Provide is called
func TestProvide_K8SecretDockerConfigJson_ReturnsExpected(t *testing.T) {
	var testSecret core.Secret
	js := `{
		"auths": {
			"index.docker.io": {
				"auth": "am9lam9lOmhlbGxv"
			}
		}
	}`
	testSecret.Data = make(map[string][]byte)
	testSecret.Data[core.DockerConfigJsonKey] = []byte(js)
	testSecret.Type = core.SecretTypeDockerConfigJson

	var secretMap = make(map[string]*core.Secret)
	secretMap["index.docker.io"] = &testSecret
	var k8secretprovder k8SecretAuthProvider
	k8secretprovder.secrets = secretMap

	authConfig, err := k8secretprovder.Provide("index.docker.io/test-artifact:v1")
	if err != nil {
		t.Fatalf("provide failed to get credential with err %v", err)
	}

	if authConfig.Username != "joejoe" || authConfig.Password != "hello" {
		t.Fatalf("provide returned incorrect credentials (username: %s, password: %s)", authConfig.Username, authConfig.Password)
	}
}

// Checks K8 DockerCfg Secret is properly extracted and
// credentials returned when Provide is called
func TestProvide_K8SecretDockerCfg_ReturnsExpected(t *testing.T) {
	var testSecret core.Secret
	js := `{
		"index.docker.io": {
			"auth": "am9lam9lOmhlbGxv"
		}
	}`
	testSecret.Data = make(map[string][]byte)
	testSecret.Data[core.DockerConfigKey] = []byte(js)
	testSecret.Type = core.SecretTypeDockercfg

	var secretMap = make(map[string]*core.Secret)
	secretMap["index.docker.io"] = &testSecret
	var k8secretprovder k8SecretAuthProvider
	k8secretprovder.secrets = secretMap

	authConfig, err := k8secretprovder.Provide("index.docker.io/test-artifact:v1")
	if err != nil {
		t.Fatalf("provide failed to get credential with err %v", err)
	}

	if authConfig.Username != "joejoe" || authConfig.Password != "hello" {
		t.Fatalf("provide returned incorrect credentials (username: %s, password: %s)", authConfig.Username, authConfig.Password)
	}
}

// Checks an error is returned for unsupported secret type
func TestProvide_K8SecretUnsupportedType_ReturnsExpected(t *testing.T) {
	var testSecret core.Secret
	testSecret.Type = core.SecretTypeOpaque
	expectedError := fmt.Errorf("secret with unsupported type %s provided", core.SecretTypeOpaque)

	var secretMap = make(map[string]*core.Secret)
	secretMap["test.ghcr.io"] = &testSecret
	var k8secretprovder k8SecretAuthProvider
	k8secretprovder.secrets = secretMap

	_, err := k8secretprovder.Provide("test.ghcr.io/test-artifact:v1")
	if err == nil {
		t.Fatalf("expected error: %s but got nil", expectedError)
	}
}
