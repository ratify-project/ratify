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

package cosign

import (
	"context"
	"testing"

	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/verifier/config"
)

// TestCreate tests the Create function of the cosign verifier
func TestCreate(t *testing.T) {
	tests := []struct {
		name    string
		config  config.VerifierConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: config.VerifierConfig{
				"name":          "test",
				"artifactTypes": "testtype",
			},
			wantErr: false,
		},
		{
			name: "missing name of verifier",
			config: config.VerifierConfig{
				"invalid": "test",
			},
			wantErr: true,
		},
		{
			name: "invalid config",
			config: config.VerifierConfig{
				"name": "test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verifierFactory := cosignVerifierFactory{}
			_, err := verifierFactory.Create("", tt.config, "", "test-namespace")
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

// TestName tests the Name function of the cosign verifier
func TestName(t *testing.T) {
	verifierFactory := cosignVerifierFactory{}
	name := "test"
	validConfig := config.VerifierConfig{
		"name":          name,
		"artifactTypes": "testtype",
	}
	cosignVerifier, err := verifierFactory.Create("", validConfig, "", "test-namespace")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if cosignVerifier.Name() != name {
		t.Errorf("Name() = %v, want %v", cosignVerifier.Name(), name)
	}
}

// TestType tests the Type function of the cosign verifier
func TestType(t *testing.T) {
	verifierFactory := cosignVerifierFactory{}
	validConfig := config.VerifierConfig{
		"name":          "test",
		"artifactTypes": "testtype",
	}
	cosignVerifier, err := verifierFactory.Create("", validConfig, "", "test-namespace")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if cosignVerifier.Type() != "cosign" {
		t.Errorf("Type() = %v, want %v", cosignVerifier.Type(), "cosign")
	}
}

// TestCanVerify tests the CanVerify function of the cosign verifier
func TestCanVerify(t *testing.T) {
	tc := []struct {
		name             string
		artifactTypes    string
		descArtifactType string
		want             bool
		shouldError      bool
	}{
		{
			name:             "valid artifact type",
			artifactTypes:    "testtype",
			descArtifactType: "testtype",
			want:             true,
		},
		{
			name:             "all artifact types",
			artifactTypes:    "*",
			descArtifactType: "testtype",
			want:             true,
		},
		{
			name:             "non matching artifact type",
			artifactTypes:    "testtype",
			descArtifactType: "othertype",
			want:             false,
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			verifierFactory := cosignVerifierFactory{}
			validConfig := config.VerifierConfig{
				"name":          "test",
				"artifactTypes": tt.artifactTypes,
			}
			cosignVerifier, err := verifierFactory.Create("", validConfig, "", "test-namespace")
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}
			result := cosignVerifier.CanVerify(context.Background(), ocispecs.ReferenceDescriptor{ArtifactType: tt.descArtifactType})
			if result != tt.want {
				t.Errorf("CanVerify() = %v, want %v", result, tt.want)
			}
		})
	}
}

// TestGetNestedReferences tests the GetNestedReferences function of the cosign verifier
func TestGetNestedReferences(t *testing.T) {
	verifierFactory := cosignVerifierFactory{}
	validConfig := config.VerifierConfig{
		"name":                "test",
		"artifactTypes":       "testtype",
		"nestedArtifactTypes": []string{"nested-artifact-type"},
	}
	cosignVerifier, err := verifierFactory.Create("", validConfig, "", "test-namespace")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	refs := cosignVerifier.GetNestedReferences()
	if len(refs) != 1 {
		t.Fatalf("GetNestedReferences() = %v, want 1", refs)
	}
	if refs[0] != "nested-artifact-type" {
		t.Errorf("GetNestedReferences() = %v, want nested-artifact-type", refs)
	}
}

// TestParseVerifierConfig tests the parseVerifierConfig function
func TestParseVerifierConfig(t *testing.T) {
	tc := []struct {
		name     string
		config   config.VerifierConfig
		wantType string
		wantErr  bool
	}{
		{
			name: "valid config",
			config: config.VerifierConfig{
				"name":                "test",
				"artifactTypes":       "testtype",
				"type":                "stuff",
				"nestedArtifactTypes": []string{"nested-artifact-type"},
				"key":                 "testkey_path",
			},
			wantType: "stuff",
			wantErr:  false,
		},
		{
			name: "missing type",
			config: config.VerifierConfig{
				"name":                "test",
				"artifactTypes":       "testtype",
				"nestedArtifactTypes": []string{"nested-artifact-type"},
				"key":                 "testkey_path",
			},
			wantType: "test",
			wantErr:  false,
		},
		{
			name: "missing name",
			config: config.VerifierConfig{
				"artifactTypes":       "testtype",
				"type":                "stuff",
				"nestedArtifactTypes": []string{"nested-artifact-type"},
				"key":                 "testkey_path",
			},
			wantType: "",
			wantErr:  true,
		},
		{
			name: "missing artifactTypes",
			config: config.VerifierConfig{
				"name":                "test",
				"type":                "stuff",
				"nestedArtifactTypes": []string{"nested-artifact-type"},
				"key":                 "testkey_path",
			},
			wantType: "",
			wantErr:  true,
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			pluginConfig, err := parseVerifierConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseVerifierConfig() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && pluginConfig.Type != tt.wantType {
				t.Errorf("Type() = %v, want %v", pluginConfig.Type, tt.wantType)
			}
		})
	}
}
