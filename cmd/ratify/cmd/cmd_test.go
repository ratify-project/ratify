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

package cmd

import (
	"strings"
	"testing"
)

const (
	configFilePath = "../../../config/config.json"
	subject        = "localhost:5000/net-monitor:v1"
	storeName      = "oras"
	digest         = "sha256:17490f904cf278d4314a1ccba407fc8fd00fb45303589b8cc7f5174ac35554f4"
)

func TestVerify(t *testing.T) {
	err := verify((verifyCmdOptions{
		subject:        subject,
		artifactTypes:  []string{""},
		configFilePath: configFilePath,
	}))

	// TODO: make ratify cli more unit testable
	// unit test should not have dependency for real image
	if !strings.Contains(err.Error(), "PLUGIN_NOT_FOUND") {
		t.Fatalf("expected containing: %s, but got: %s", "PLUGIN_NOT_FOUND", err.Error())
	}
}

func TestDiscover(t *testing.T) {
	err := discover((discoverCmdOptions{
		subject:        subject,
		artifactTypes:  []string{""},
		configFilePath: configFilePath,
	}))

	// TODO: make ratify cli more unit testable
	// unit test should not need to resolve real image
	if !strings.Contains(err.Error(), "REFERRER_STORE_FAILURE") {
		t.Errorf("expected containing: %s, but got: %s", "REFERRER_STORE_FAILURE", err.Error())
	}
}

func TestShowRefManifest(t *testing.T) {
	err := showRefManifest((referrerCmdOptions{
		subject:        subject,
		configFilePath: configFilePath,
		storeName:      storeName,
		digest:         digest,
	}))

	// TODO: make ratify cli more unit testable
	// unit test should not need to resolve real image
	if !strings.Contains(err.Error(), "failed to resolve subject descriptor") {
		t.Errorf("error expected")
	}

	// validate show blob returns error
	err = showBlob((referrerCmdOptions{
		subject:        subject,
		configFilePath: configFilePath,
		storeName:      storeName,
		digest:         "invalid-digest",
	}))

	if !strings.Contains(err.Error(), "the digest of the subject is invalid") {
		t.Errorf("error expected")
	}
}
