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
	"fmt"
	"testing"

	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/verifier/config"
	imgspec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sigstore/cosign/v2/pkg/oci/static"
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

// TestStaticLayerOpts tests the staticLayerOpts function
func TestStaticLayerOpts(test *testing.T) {
	tc := []struct {
		name    string
		desc    imgspec.Descriptor
		wantLen int
		wantErr bool
	}{
		{
			name: "valid config",
			desc: imgspec.Descriptor{
				MediaType: "application/vnd.oci.image.layer.v1.tar",
				Annotations: map[string]string{
					static.CertificateAnnotationKey: "testcert",
					static.ChainAnnotationKey:       "testchain",
					static.BundleAnnotationKey:      "{\"SignedEntryTimestamp\":\"MEUCIQDj21LM7GkNM3SlaodBrUgZBFF7gFCgI1u/bE82kCyzBgIgB7Oqk/vahZevqUTHqo7JFo02yc2zawTTw3gMwwf0En8=\",\"Payload\":{\"body\":\"eyJhcGlWZXJzaW9uIjoiMC4wLjEiLCJraW5kIjoiaGFzaGVkcmVrb3JkIiwic3BlYyI6eyJkYXRhIjp7Imhhc2giOnsiYWxnb3JpdGhtIjoic2hhMjU2IiwidmFsdWUiOiI4MzdhMjkyNGZlZjgwYTVhNmJiOGQ4NzExZTY2NTE4YzMwOTJkZTg4ODhlYTg4ODBhOTMyM2I5ODUwNjkzMzM5In19LCJzaWduYXR1cmUiOnsiY29udGVudCI6Ik1FUUNJQXFueXVsUTVoZW1KaStQZ0EvdHgyVEFqM2xSdlJlYkFmSWtZYnM1NlZhckFpQnJoS3BaQ3Fzb2M2NGdzdllEZUxFWGRyb291U0RvQjRrZ2dJUEZ6MEFpSnc9PSIsInB1YmxpY0tleSI6eyJjb250ZW50IjoiTFMwdExTMUNSVWRKVGlCRFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVTXhla05EUVd3eVowRjNTVUpCWjBsVldXOUlibWcyY25GQk0yZFVSSE5TTTFaUmRuSmFkazlQT0RWTmQwTm5XVWxMYjFwSmVtb3dSVUYzVFhjS1RucEZWazFDVFVkQk1WVkZRMmhOVFdNeWJHNWpNMUoyWTIxVmRWcEhWakpOVWpSM1NFRlpSRlpSVVVSRmVGWjZZVmRrZW1SSE9YbGFVekZ3WW01U2JBcGpiVEZzV2tkc2FHUkhWWGRJYUdOT1RXcE5kMDVFUlROTmFrRXhUVlJGTTFkb1kwNU5hazEzVGtSRk0wMXFSWGROVkVVelYycEJRVTFHYTNkRmQxbElDa3R2V2tsNmFqQkRRVkZaU1V0dldrbDZhakJFUVZGalJGRm5RVVV3VFRNd2FWTndaVUU1WVhGMk1VOUJVVEp6YlRKbVQwSkpSbUpMWVhOck1UVktOVTRLU25SSkswRTJUR2RFU1V0WVdYcEJTRXB0ZDBabFVGUkRZbHB6V2t3cldHeDZaVTlCTmxFelZYWXhPR1pFUWpoUFMyRlBRMEZZZDNkblowWTBUVUUwUndwQk1WVmtSSGRGUWk5M1VVVkJkMGxJWjBSQlZFSm5UbFpJVTFWRlJFUkJTMEpuWjNKQ1owVkdRbEZqUkVGNlFXUkNaMDVXU0ZFMFJVWm5VVlUwV1hsMUNqZzJMMjV2Um1Vclptb3ZkbFIwSzBKSlJ6Qk1WelEwZDBoM1dVUldVakJxUWtKbmQwWnZRVlV6T1ZCd2VqRlphMFZhWWpWeFRtcHdTMFpYYVhocE5Ga0tXa1E0ZDBwM1dVUldVakJTUVZGSUwwSkNNSGRITkVWYVdWZDBhR015YUhKak1teDFXakpvYUdKRWF6UlJSMlIwV1Zkc2MweHRUblppVkVGelFtZHZjZ3BDWjBWRlFWbFBMMDFCUlVKQ1FqVnZaRWhTZDJONmIzWk1NbVJ3WkVkb01WbHBOV3BpTWpCMllrYzVibUZYTkhaaU1rWXhaRWRuZDB4bldVdExkMWxDQ2tKQlIwUjJla0ZDUTBGUlowUkNOVzlrU0ZKM1kzcHZka3d5WkhCa1IyZ3hXV2sxYW1JeU1IWmlSemx1WVZjMGRtSXlSakZrUjJkM1oxbHJSME5wYzBjS1FWRlJRakZ1YTBOQ1FVbEZaWGRTTlVGSVkwRmtVVVJrVUZSQ2NYaHpZMUpOYlUxYVNHaDVXbHA2WTBOdmEzQmxkVTQwT0hKbUswaHBia3RCVEhsdWRRcHFaMEZCUVZsbFVTOUlhMkZCUVVGRlFYZENSMDFGVVVOSlIyTm5UVFl2YVZGWldIUlVkemhoWVUxSUwweDRhQzlYYjA4eGVISlFOVUZoY2xOV1JEaFlDa2t3YURSQmFVSk5jRU5NTkN0RlMxbHNPWGQ2TnpBd1JuUXdWSEIxVDNoVVJFVnhhR3hFVEZsVmNIWnFjR1prUW5aRVFVdENaMmR4YUd0cVQxQlJVVVFLUVhkT2IwRkVRbXhCYWtFeFpuZDVaVkowVnpGcU4yUllNbHA2UzA1aVluRjNjVU0zWVZjNWFXTlhlRE40U1dsRU1IQlZhVGRLWVRodWRWbG5XR3BJYUFwWVZVWnBRMng2UjNkS1kwTk5VVVEzV21GTVNERkxUekZwVTNwU1EwVkZaRGhhUm1KaFprMXhjSGc1VTBsVFVHaGxSVFJsYkdWV1NDdEhjbkZ5Ylc5WkNrMVBUbHB2TVhCa1ltOVZTMnh2TUQwS0xTMHRMUzFGVGtRZ1EwVlNWRWxHU1VOQlZFVXRMUzB0TFFvPSJ9fX19\",\"integratedTime\":1681764685,\"logIndex\":18215184,\"logID\":\"c0d23d6ad406973f9559f3ba2d1ca01f84147d8ffc5b8445c224f98b9591801d\"}}",
				},
			},
			wantLen: 3,
			wantErr: false,
		},
		{
			name: "incorrect rekor bundle",
			desc: imgspec.Descriptor{
				MediaType: "application/vnd.oci.image.layer.v1.tar",
				Annotations: map[string]string{
					static.CertificateAnnotationKey: "testcert",
					static.ChainAnnotationKey:       "testchain",
					static.BundleAnnotationKey:      "invalidbundle",
				},
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tc {
		test.Run(tt.name, func(test *testing.T) {
			layerOpts, err := staticLayerOpts(tt.desc)
			if (err != nil) != tt.wantErr {
				test.Errorf("staticLayerOpts() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && len(layerOpts) != tt.wantLen {
				test.Errorf("staticLayerOpts() = %v, want %v", len(layerOpts), tt.wantLen)
			}
		})
	}
}

// TestErrorToVerifyResult tests the errorToVerifyResult function
func TestErrorToVerifyResult(t *testing.T) {
	verifierResult := errorToVerifyResult("test", "cosign", fmt.Errorf("test error"))
	if verifierResult.IsSuccess {
		t.Errorf("errorToVerifyResult() = %v, want %v", verifierResult.IsSuccess, false)
	}
	if verifierResult.Name != "test" {
		t.Errorf("errorToVerifyResult() = %v, want %v", verifierResult.Name, "test")
	}
	if verifierResult.Type != "cosign" {
		t.Errorf("errorToVerifyResult() = %v, want %v", verifierResult.Type, "cosign")
	}
	if verifierResult.Message != "cosign verification failed: test error" {
		t.Errorf("errorToVerifyResult() = %v, want %v", verifierResult.Message, "cosign verification failed: test error")
	}
}
