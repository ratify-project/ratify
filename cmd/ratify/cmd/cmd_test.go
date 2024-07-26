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
)

func TestVerify(t *testing.T) {
	err := verify((verifyCmdOptions{
		subject:        subject,
		artifactTypes:  []string{""},
		configFilePath: configFilePath,
	}))

	if !strings.Contains(err.Error(), "plugin not found") {
		t.Errorf("error expected")
	}
}

func TestDiscover(t *testing.T) {
	// validate discover command does not crash
	err := discover((discoverCmdOptions{
		subject:        subject,
		artifactTypes:  []string{""},
		configFilePath: configFilePath,
	}))

	if !strings.Contains(err.Error(), "referrer store failure") {
		t.Errorf("error expected")
	}
}

func TestShowRefManifest(t *testing.T) {
	// validate discover command does not crash
	err := showRefManifest((referrerCmdOptions{
		subject:        subject,
		configFilePath: configFilePath,
	}))

	if !strings.Contains(err.Error(), "store name parameter is required") {
		t.Errorf("error expected")
	}
}
