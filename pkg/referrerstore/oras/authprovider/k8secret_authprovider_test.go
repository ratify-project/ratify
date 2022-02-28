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

	secretList := []*core.Secret{&testSecret}
	var k8secretprovder k8SecretAuthProvider
	k8secretprovder.secrets = secretList

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

	secretList := []*core.Secret{&testSecret}
	var k8secretprovder k8SecretAuthProvider
	k8secretprovder.secrets = secretList

	authConfig, err := k8secretprovder.Provide("index.docker.io/test-artifact:v1")
	if err != nil {
		t.Fatalf("provide failed to get credential with err %v", err)
	}

	if authConfig.Username != "joejoe" || authConfig.Password != "hello" {
		t.Fatalf("provide returned incorrect credentials (username: %s, password: %s)", authConfig.Username, authConfig.Password)
	}
}

// // Checks an error is returned for non-existent registry credential
func TestProvide_K8SecretNonExistentRegistry_ReturnsExpected(t *testing.T) {
	var testSecret core.Secret
	testArtifact := "nonexistent.ghcr.io/test-artifact:v1"
	expectedErr := fmt.Errorf("could not find credentials for %s", testArtifact)
	js := `{
		"index.docker.io": {
			"auth": "am9lam9lOmhlbGxv"
		}
	}`
	testSecret.Data = make(map[string][]byte)
	testSecret.Data[core.DockerConfigKey] = []byte(js)
	testSecret.Type = core.SecretTypeDockercfg

	secretList := []*core.Secret{&testSecret}
	var k8secretprovder k8SecretAuthProvider
	k8secretprovder.secrets = secretList

	_, err := k8secretprovder.Provide(testArtifact)
	if err.Error() != expectedErr.Error() {
		t.Fatalf("expected err: %s, but got err: %s", expectedErr, err)
	}
}
