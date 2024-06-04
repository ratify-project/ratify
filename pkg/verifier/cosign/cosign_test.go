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
	"crypto"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/keymanagementprovider"
	"github.com/deislabs/ratify/pkg/keymanagementprovider/azurekeyvault"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore/mocks"
	"github.com/deislabs/ratify/pkg/verifier/config"
	"github.com/opencontainers/go-digest"
	imgspec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sigstore/cosign/v2/pkg/cosign"
	"github.com/sigstore/cosign/v2/pkg/oci/static"
	"github.com/sigstore/rekor/pkg/generated/client"
	"github.com/sigstore/sigstore/pkg/cryptoutils"
	"github.com/sigstore/sigstore/pkg/signature"
)

const (
	ratifySampleImageRef string = "ghcr.io/deislabs/ratify:v1"
	testIdentity         string = "test-identity"
	testIssuer           string = "https://test-issuer.com"
)

type mockNoOpVerifier struct{}

func (m *mockNoOpVerifier) PublicKey(_ ...signature.PublicKeyOption) (crypto.PublicKey, error) {
	return nil, nil
}

func (m *mockNoOpVerifier) VerifySignature(_, _ io.Reader, _ ...signature.VerifyOption) error {
	return nil
}

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
				"trustPolicies": []TrustPolicyConfig{
					{
						Name:    "test",
						Keyless: KeylessConfig{CertificateIdentity: testIdentity, CertificateOIDCIssuer: testIssuer},
						Scopes:  []string{"*"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid legacy config",
			config: config.VerifierConfig{
				"name":          "test",
				"artifactTypes": "testtype",
				"key":           "testkey_path",
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
		{
			name: "valid trust policies config with no legacy config or trust policies",
			config: config.VerifierConfig{
				"name":          "test",
				"artifactTypes": "testtype",
				"trustPolicies": []TrustPolicyConfig{},
			},
			wantErr: false,
		},
		{
			name: "invalid config with legacy and trust policies",
			config: config.VerifierConfig{
				"name":          "test",
				"artifactTypes": "testtype",
				"trustPolicies": []TrustPolicyConfig{
					{
						Name:    "test",
						Keyless: KeylessConfig{CertificateIdentity: testIdentity, CertificateOIDCIssuer: testIssuer},
						Scopes:  []string{"*"},
					},
				},
				"key": "testkey_path",
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
		"trustPolicies": []TrustPolicyConfig{
			{
				Name:    "test",
				Keyless: KeylessConfig{CertificateIdentity: testIdentity, CertificateOIDCIssuer: testIssuer},
				Scopes:  []string{"*"},
			},
		},
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
		"trustPolicies": []TrustPolicyConfig{
			{
				Name:    "test",
				Keyless: KeylessConfig{CertificateIdentity: testIdentity, CertificateOIDCIssuer: testIssuer},
				Scopes:  []string{"*"},
			},
		},
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
				"trustPolicies": []TrustPolicyConfig{
					{
						Name:    "test",
						Keyless: KeylessConfig{CertificateIdentity: testIdentity, CertificateOIDCIssuer: testIssuer},
						Scopes:  []string{"*"},
					},
				},
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
		"name":          "test",
		"artifactTypes": "testtype",
		"trustPolicies": []TrustPolicyConfig{
			{
				Name:    "test",
				Keyless: KeylessConfig{CertificateIdentity: testIdentity, CertificateOIDCIssuer: testIssuer},
				Scopes:  []string{"*"},
			},
		},
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

// TestDecodeASN1Signature tests the decodeASN1Signature function
func TestDecodeASN1Signature(t *testing.T) {
	tc := []struct {
		name             string
		sigBytes         []byte
		expectedSigBytes []byte
		wantErr          bool
	}{
		{
			name:             "invalid not asn1",
			sigBytes:         []byte("test"),
			expectedSigBytes: []byte("test"),
			wantErr:          false,
		},
		{
			name:             "valid asn1",
			sigBytes:         []byte("0E\x02!\x00\xb4\xd7R\xf0\xee\x11ձ\x9f\rtsog\x99\xa1\x86L=\x04\x92\u07b8\xb7\xa1\x94Mj\xfe\xd2\xda\x02\x02 \x027|~q\xcb2\xaf\xd1\xddx;\x04\xed\r\x9a\x9a\x03\xa9\x84\x8cu\xba\x1a\x9eFb\xc2h\x7fk\xc3"),
			expectedSigBytes: []byte("\xb4\xd7R\xf0\xee\x11ձ\x9f\rtsog\x99\xa1\x86L=\x04\x92\u07b8\xb7\xa1\x94Mj\xfe\xd2\xda\x02\x027|~q\xcb2\xaf\xd1\xddx;\x04\xed\r\x9a\x9a\x03\xa9\x84\x8cu\xba\x1a\x9eFb\xc2h\x7fk\xc3"),
			wantErr:          false,
		},
		{
			name:             "invalid r",
			sigBytes:         []byte("0E\x03!\x00\xb4\xd7R\xf0\xee\x11ձ\x9f\rtsog\x99\xa1\x86L=\x04\x92\u07b8\xb7\xa1\x94Mj\xfe\xd2\xda\x02\x02 \x027|~q\xcb2\xaf\xd1\xddx;\x04\xed\r\x9a\x9a\x03\xa9\x84\x8cu\xba\x1a\x9eFb\xc2h\x7fk\xc3"),
			expectedSigBytes: nil,
			wantErr:          true,
		},
		{
			name:             "invalid s",
			sigBytes:         []byte("0E\x02!\x00\xb4\xd7R\xf0\xee\x11ձ\x9f\rtsog\x99\xa1\x86L=\x04\x92\u07b8\xb7\xa1\x94Mj\xfe\xd2\xda\x02\x03 \x027|~q\xcb2\xaf\xd1\xddx;\x04\xed\r\x9a\x9a\x03\xa9\x84\x8cu\xba\x1a\x9eFb\xc2h\x7fk\xc3"),
			expectedSigBytes: nil,
			wantErr:          true,
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			sigBytes, err := decodeASN1Signature(tt.sigBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeASN1Signature() error = %v, wantErr %v", err, tt.wantErr)
			}
			if sigBytes != nil && !slices.Equal(tt.expectedSigBytes, sigBytes) {
				t.Errorf("decodeASN1Signature() = %v, want %v", sigBytes, tt.expectedSigBytes)
			}
		})
	}
}

func TestGetKeysMaps_Success(t *testing.T) {
	trustPolicy := &mockTrustPolicy{}
	_, _, err := getKeyMapOptsDefault(context.Background(), trustPolicy, "gatekeeper-system")
	if err != nil {
		t.Errorf("getKeysMaps() error = %v, wantErr %v", err, false)
	}
}

func TestGetKeysMaps_FailingCosignOpts(t *testing.T) {
	trustPolicy := &mockTrustPolicy{shouldErrCosignOpts: true}
	_, _, err := getKeyMapOptsDefault(context.Background(), trustPolicy, "gatekeeper-system")
	if err == nil {
		t.Errorf("getKeysMaps() error = %v, wantErr %v", err, true)
	}
}

func TestGetKeysMaps_FailingGetKeys(t *testing.T) {
	trustPolicy := &mockTrustPolicy{shouldErrKeys: true}
	_, _, err := getKeyMapOptsDefault(context.Background(), trustPolicy, "gatekeeper-system")
	if err == nil {
		t.Errorf("getKeysMaps() error = %v, wantErr %v", err, true)
	}
}

// TestVerifyInternal tests the verifyInternal function of the cosign verifier
// it also tests the processAKVSignature function implicitly
func TestVerifyInternal(t *testing.T) {
	cosignMediaType := "application/vnd.dev.cosign.simplesigning.v1+json"
	validSignatureBlob := []byte("test")
	subjectDigest := digest.Digest("sha256:5678")
	testRefDigest := digest.Digest("sha256:1234")
	blobDigest := digest.Digest("valid blob")
	//nolint:gosec // this is a test key
	invalidRsaPrivKey, _ := rsa.GenerateKey(rand.Reader, 1024)
	invalidRsaPubKey := invalidRsaPrivKey.Public()
	invalidECPrivKey, _ := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	invalidECPubKey := invalidECPrivKey.Public()

	validRSA256PubKey, err := cryptoutils.UnmarshalPEMToPublicKey([]byte(`-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtHGXFzi1W93vQ88EwzmI
IhTXMYpcffQmmHYLgjkeLWL4SQ7DJEyq4j+Yz994lq0B4nCT9EaLkXYSMZhfYuHg
y+2kMkh+1eNUtjGVJBkHc5iz7YR9OaIDlY36TnKlk0HfyBrjNrlwyodD6no/2LCf
6FmGT6mVIaE/fyrxN3ZCHzfcw5LaGgHRt+91NJa5PnQCxjXUfyabHbHehgNLjjpn
kwCW3GGs56cOMQowHsLrlwQnXAq5wvAueRz3Ommz+DPnVXUSV+vfYDt56oggX386
LOe8VCiwi4T9IIuWlKIi+AuIm8aQ+11o9LjpvDqFD1rJU/KMFhczA4Kj0fRM7Ulg
ewIDAQAB
-----END PUBLIC KEY-----`))
	if err != nil {
		t.Fatalf("error creating RSA public key: %v", err)
	}

	validRSA384PubKey, err := cryptoutils.UnmarshalPEMToPublicKey([]byte(`-----BEGIN PUBLIC KEY-----
MIIBojANBgkqhkiG9w0BAQEFAAOCAY8AMIIBigKCAYEArccnwmvTOS0iqaxiCRsD
4vixdUy2az39vL3iABjLr6Ht2NLA8dyO0NeEBb6+lfoqOl8RX29rJ0LnCaQL/wV8
BQ3idfzdeX/rzdhRoegDiZ7MDgd01ZDeocGSfOAKJ3E1Kr0+etB4UuOF2T7dcVNn
lAtZxEH6wNtW2HFoLg6bnlUuSj+9RVaP2Z0D55Bk4Un1jinB6Et81SCIuvDcMbKt
aW3Xu17mdHiscLQBOnmX86mKRP8R4Ij9TtNyEW/9WLNXHV1iJhm9TVONZkX2hRjy
o3+pPYvsZAAjyIk4AHF4BROCMA+WmyqkjnyVdEcJBi6f8NptjnS8A5jtTXIrndEq
OE1VTu44z8hcQqrIypdyF86rMsJm91F8x68clvSYyvYn15lzv720LOglFF2NrD8S
0SmxbyPB4bnRhEiyxh9ocAbGVXu+FyjrLPjTCTTnIpnTzgm/XtSqjA6104Zz3Seu
TvvdnkTLbUxqHzoFYXSksJHvOiH2U7JAay8S4CZ4KrGvAgMBAAE=
-----END PUBLIC KEY-----`))
	if err != nil {
		t.Fatalf("error creating RSA public key: %v", err)
	}

	validECDSAP256PubKey, err := cryptoutils.UnmarshalPEMToPublicKey([]byte(`-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE1ljPT4AO3Ny57S2B1a5P2LSrru5l
ewt8iyi46g8SRrasghTR8xliLiZJl3GTM3UOdUAZCiruwPC7hihLD5JcqQ==
-----END PUBLIC KEY-----`))
	if err != nil {
		t.Fatalf("error creating RSA public key: %v", err)
	}

	validECDSAP384PubKey, err := cryptoutils.UnmarshalPEMToPublicKey([]byte(`-----BEGIN PUBLIC KEY-----
MHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEFbJSMiBAtIiydUeqhMGpZBDRkZhYFu5r
zg5rpyR7WJVgDPH8++2IY9Zg1HYKB0aGqyuLK5i8bG3C8avDLXg2+Dmf35wV2Mgh
mmBwUAwwW0Uc+Nt3bDOCiB1nUsICv1ry
-----END PUBLIC KEY-----
	`))
	if err != nil {
		t.Fatalf("error creating RSA public key: %v", err)
	}

	subjectRef := common.Reference{
		Digest:   subjectDigest,
		Original: ratifySampleImageRef,
		Tag:      "v1",
	}
	refDescriptor := ocispecs.ReferenceDescriptor{
		ArtifactType: "testtype",
		Descriptor: imgspec.Descriptor{
			Digest:    testRefDigest,
			MediaType: imgspec.MediaTypeImageManifest,
		},
	}
	tc := []struct {
		name                        string
		keys                        map[PKKey]keymanagementprovider.PublicKey
		getKeysError                bool
		cosignOpts                  cosign.CheckOpts
		store                       *mocks.MemoryTestStore
		expectedResultMessagePrefix string
	}{
		{
			name:                        "get keys error",
			keys:                        map[PKKey]keymanagementprovider.PublicKey{},
			getKeysError:                true,
			store:                       &mocks.MemoryTestStore{},
			expectedResultMessagePrefix: "cosign verification failed: error",
		},
		{
			name:                        "manifest fetch error",
			keys:                        map[PKKey]keymanagementprovider.PublicKey{},
			getKeysError:                false,
			store:                       &mocks.MemoryTestStore{},
			expectedResultMessagePrefix: "cosign verification failed: failed to get reference manifest",
		},
		{
			name:         "incorrect reference manifest media type error",
			keys:         map[PKKey]keymanagementprovider.PublicKey{},
			getKeysError: false,
			store: &mocks.MemoryTestStore{
				Manifests: map[digest.Digest]ocispecs.ReferenceManifest{
					testRefDigest: {
						MediaType: "invalid",
					},
				},
			},
			expectedResultMessagePrefix: "cosign verification failed: reference manifest is not an image",
		},
		{
			name:         "failed subject descriptor fetch",
			keys:         map[PKKey]keymanagementprovider.PublicKey{},
			getKeysError: false,
			store: &mocks.MemoryTestStore{
				Manifests: map[digest.Digest]ocispecs.ReferenceManifest{
					testRefDigest: {
						MediaType: refDescriptor.MediaType,
					},
				},
			},
			expectedResultMessagePrefix: "cosign verification failed: failed to create subject hash",
		},
		{
			name:         "failed to fetch blob",
			keys:         map[PKKey]keymanagementprovider.PublicKey{},
			getKeysError: false,
			store: &mocks.MemoryTestStore{
				Manifests: map[digest.Digest]ocispecs.ReferenceManifest{
					testRefDigest: {
						MediaType: refDescriptor.MediaType,
						Blobs: []imgspec.Descriptor{
							{
								Digest: digest.Digest("nonexistent blob hash"),
							},
						},
					},
				},
				Subjects: map[digest.Digest]*ocispecs.SubjectDescriptor{
					subjectDigest: {
						Descriptor: imgspec.Descriptor{
							Digest:    subjectDigest,
							MediaType: imgspec.MediaTypeImageManifest,
						},
					},
				},
			},
			expectedResultMessagePrefix: "cosign verification failed: failed to get blob content",
		},
		{
			name: "invalid key type for AKV",
			keys: map[PKKey]keymanagementprovider.PublicKey{
				{Provider: "test"}: {Key: &ecdh.PublicKey{}, ProviderType: azurekeyvault.ProviderName},
			},
			getKeysError: false,
			store: &mocks.MemoryTestStore{
				Manifests: map[digest.Digest]ocispecs.ReferenceManifest{
					testRefDigest: {
						MediaType: refDescriptor.MediaType,
						Blobs: []imgspec.Descriptor{
							{
								Digest: blobDigest,
							},
						},
					},
				},
				Subjects: map[digest.Digest]*ocispecs.SubjectDescriptor{
					subjectDigest: {
						Descriptor: imgspec.Descriptor{
							Digest:    subjectDigest,
							MediaType: imgspec.MediaTypeImageManifest,
						},
					},
				},
				Blobs: map[digest.Digest][]byte{
					blobDigest: validSignatureBlob,
				},
			},
			expectedResultMessagePrefix: "cosign verification failed: failed to verify with keys: failed to process AKV signature: unsupported public key type",
		},
		{
			name: "invalid RSA key size for AKV",
			keys: map[PKKey]keymanagementprovider.PublicKey{
				{Provider: "test"}: {Key: invalidRsaPubKey, ProviderType: azurekeyvault.ProviderName},
			},
			getKeysError: false,
			store: &mocks.MemoryTestStore{
				Manifests: map[digest.Digest]ocispecs.ReferenceManifest{
					testRefDigest: {
						MediaType: refDescriptor.MediaType,
						Blobs: []imgspec.Descriptor{
							{
								Digest: blobDigest,
							},
						},
					},
				},
				Subjects: map[digest.Digest]*ocispecs.SubjectDescriptor{
					subjectDigest: {
						Descriptor: imgspec.Descriptor{
							Digest:    subjectDigest,
							MediaType: imgspec.MediaTypeImageManifest,
						},
					},
				},
				Blobs: map[digest.Digest][]byte{
					blobDigest: validSignatureBlob,
				},
			},
			expectedResultMessagePrefix: "cosign verification failed: failed to verify with keys: failed to process AKV signature: RSA key check: unsupported key size",
		},
		{
			name: "invalid ECDSA curve type for AKV",
			keys: map[PKKey]keymanagementprovider.PublicKey{
				{Provider: "test"}: {Key: invalidECPubKey, ProviderType: azurekeyvault.ProviderName},
			},
			getKeysError: false,
			store: &mocks.MemoryTestStore{
				Manifests: map[digest.Digest]ocispecs.ReferenceManifest{
					testRefDigest: {
						MediaType: refDescriptor.MediaType,
						Blobs: []imgspec.Descriptor{
							{
								Digest: blobDigest,
							},
						},
					},
				},
				Subjects: map[digest.Digest]*ocispecs.SubjectDescriptor{
					subjectDigest: {
						Descriptor: imgspec.Descriptor{
							Digest:    subjectDigest,
							MediaType: imgspec.MediaTypeImageManifest,
						},
					},
				},
				Blobs: map[digest.Digest][]byte{
					blobDigest: validSignatureBlob,
				},
			},
			expectedResultMessagePrefix: "cosign verification failed: failed to verify with keys: failed to process AKV signature: ECDSA key check: unsupported key curve",
		},
		{
			name: "valid hash: 256 keysize: 2048 RSA key AKV",
			keys: map[PKKey]keymanagementprovider.PublicKey{
				{Provider: "test"}: {Key: validRSA256PubKey, ProviderType: azurekeyvault.ProviderName},
			},
			cosignOpts: cosign.CheckOpts{
				IgnoreSCT:  true,
				IgnoreTlog: true,
			},
			getKeysError: false,
			store: &mocks.MemoryTestStore{
				Manifests: map[digest.Digest]ocispecs.ReferenceManifest{
					testRefDigest: {
						MediaType: refDescriptor.MediaType,
						Blobs: []imgspec.Descriptor{
							{
								Size:      267,
								Digest:    "sha256:6e1ffef2ba058cda5d1aa7ed792cb1e63b4207d8195a469bee1b5fc662cd9b70",
								MediaType: cosignMediaType,
								Annotations: map[string]string{
									static.SignatureAnnotationKey: "j6VNQ+Z3BqLeM75WM8WKnJqtwR7Kv21BwURHLmK6S05gCV/JntSbVthNVKoNY3906NMqmfZDlP/QuUOQt7Fxq2ivixw1xKa1KlE+ydW951GyMysaZx36U08Wmfyqt6dbgXMU6/nQE8oxG855rfywvE+MAmIJ+u1ktPbU+HoXEPP8yNUyK4gY/JAopQVEcktGAqFAbT49LzlE3FTJQNE6WryCQy5GiaM/1qdKzQi9GQb2g20Vxg6+e4AuxogAs+bzexoA4J5bUkDAkE/PDIXNz2EgjB0o7zK1NQEDiLNRq7fafTY5G56vXtltuMWOzCGnLMXbk4f3K9wKXF++7h4I3w==",
								},
							},
						},
					},
				},
				Subjects: map[digest.Digest]*ocispecs.SubjectDescriptor{
					subjectDigest: {
						Descriptor: imgspec.Descriptor{
							Digest:    subjectDigest,
							MediaType: imgspec.MediaTypeImageManifest,
						},
					},
				},
				Blobs: map[digest.Digest][]byte{
					"sha256:6e1ffef2ba058cda5d1aa7ed792cb1e63b4207d8195a469bee1b5fc662cd9b70": []byte(`{"critical":{"identity":{"docker-reference":"artifactstest.azurecr.io/4-15-24/cosign/hello-world"},"image":{"docker-manifest-digest":"sha256:d37ada95d47ad12224c205a938129df7a3e52345828b4fa27b03a98825d1e2e7"},"type":"cosign container image signature"},"optional":null}`),
				},
			},
			expectedResultMessagePrefix: "cosign verification success",
		},
		{
			name: "valid hash: 256 keysize: 3072 RSA key",
			keys: map[PKKey]keymanagementprovider.PublicKey{
				{Provider: "test"}: {Key: validRSA384PubKey},
			},
			cosignOpts: cosign.CheckOpts{
				IgnoreSCT:  true,
				IgnoreTlog: true,
			},
			getKeysError: false,
			store: &mocks.MemoryTestStore{
				Manifests: map[digest.Digest]ocispecs.ReferenceManifest{
					testRefDigest: {
						MediaType: refDescriptor.MediaType,
						Blobs: []imgspec.Descriptor{
							{
								Size:      267,
								Digest:    "sha256:6e1ffef2ba058cda5d1aa7ed792cb1e63b4207d8195a469bee1b5fc662cd9b70",
								MediaType: cosignMediaType,
								Annotations: map[string]string{
									static.SignatureAnnotationKey: "fP5+FQcc59WjqDAcvcgfHBZbu/FfQYh+ZjgwuEwLj/y0ku2S+rFbk8XE2gPZ4mcgT9Bceu+UMY/pYLqNI7ngkXMamYg1gzsTPrAG5DpEbApGMDiQyOlCcEFqgJbxqFOmg+HD9eSOMmibFbUh8XMt4LuyZIjmcCqJ22i8B49y8LFo6QiE64/jjhNLlRK4LvDTSUGDJ4VXW+c9y/PxbpZxtHIVyIYK82qL8P2/BuRxQ9ZVKJE1eFdz3Suz0ZIQmhkimLqQdOOxoGFcO4syjHYzfneBNvySWNxVXJCjw86DJqsDl5se+mY2Zww13DihfQX0cKSGGVfRoMgvIQOeaMNyFaCad2BQFfraqVUU5p7v0FqO6r0FU9z0ixRj81xVKJA3GPUZdF1ImcwOE4cOuQYARE6aiw78t2vrW5PRGtRPWpu+JY1+2v5m61w60G9HAozpnucWG3u9agdhwwD6VLJzPduVdnZr8t1WN8BpZs5NA3n4wkrlmRpnYtw7MqupaJQ2",
								},
							},
						},
					},
				},
				Subjects: map[digest.Digest]*ocispecs.SubjectDescriptor{
					subjectDigest: {
						Descriptor: imgspec.Descriptor{
							Digest:    subjectDigest,
							MediaType: imgspec.MediaTypeImageManifest,
						},
					},
				},
				Blobs: map[digest.Digest][]byte{
					"sha256:6e1ffef2ba058cda5d1aa7ed792cb1e63b4207d8195a469bee1b5fc662cd9b70": []byte(`{"critical":{"identity":{"docker-reference":"artifactstest.azurecr.io/4-15-24/cosign/hello-world"},"image":{"docker-manifest-digest":"sha256:d37ada95d47ad12224c205a938129df7a3e52345828b4fa27b03a98825d1e2e7"},"type":"cosign container image signature"},"optional":null}`),
				},
			},
			expectedResultMessagePrefix: "cosign verification success",
		},
		{
			name: "valid hash: 256 curve: P256 ECDSA key AKV",
			keys: map[PKKey]keymanagementprovider.PublicKey{
				{Provider: "test"}: {Key: validECDSAP256PubKey, ProviderType: azurekeyvault.ProviderName},
			},
			cosignOpts: cosign.CheckOpts{
				IgnoreSCT:  true,
				IgnoreTlog: true,
			},
			getKeysError: false,
			store: &mocks.MemoryTestStore{
				Manifests: map[digest.Digest]ocispecs.ReferenceManifest{
					testRefDigest: {
						MediaType: refDescriptor.MediaType,
						Blobs: []imgspec.Descriptor{
							{
								Size:      267,
								Digest:    "sha256:6e1ffef2ba058cda5d1aa7ed792cb1e63b4207d8195a469bee1b5fc662cd9b70",
								MediaType: cosignMediaType,
								Annotations: map[string]string{
									static.SignatureAnnotationKey: "MEYCIQDCMOtZXzsgZknsOhcv1VC7cN72xuBr16GU98bT0tXWdQIhAJp9X9jh4sIG1xhmoaYwGGkl1/8EQW7zqFUpMkEoi3s1",
								},
							},
						},
					},
				},
				Subjects: map[digest.Digest]*ocispecs.SubjectDescriptor{
					subjectDigest: {
						Descriptor: imgspec.Descriptor{
							Digest:    subjectDigest,
							MediaType: imgspec.MediaTypeImageManifest,
						},
					},
				},
				Blobs: map[digest.Digest][]byte{
					"sha256:6e1ffef2ba058cda5d1aa7ed792cb1e63b4207d8195a469bee1b5fc662cd9b70": []byte(`{"critical":{"identity":{"docker-reference":"artifactstest.azurecr.io/4-15-24/cosign/hello-world"},"image":{"docker-manifest-digest":"sha256:d37ada95d47ad12224c205a938129df7a3e52345828b4fa27b03a98825d1e2e7"},"type":"cosign container image signature"},"optional":null}`),
				},
			},
			expectedResultMessagePrefix: "cosign verification success",
		},
		{
			name: "valid hash: 256 curve: P384 ECDSA key",
			keys: map[PKKey]keymanagementprovider.PublicKey{
				{Provider: "test"}: {Key: validECDSAP384PubKey},
			},
			cosignOpts: cosign.CheckOpts{
				IgnoreSCT:  true,
				IgnoreTlog: true,
			},
			getKeysError: false,
			store: &mocks.MemoryTestStore{
				Manifests: map[digest.Digest]ocispecs.ReferenceManifest{
					testRefDigest: {
						MediaType: refDescriptor.MediaType,
						Blobs: []imgspec.Descriptor{
							{
								Size:      267,
								Digest:    "sha256:6e1ffef2ba058cda5d1aa7ed792cb1e63b4207d8195a469bee1b5fc662cd9b70",
								MediaType: cosignMediaType,
								Annotations: map[string]string{
									static.SignatureAnnotationKey: "MGUCMQC6Z7RgD2uxG5IiqKoOmrjTRVqBn+XqSjHU5oSI/RNAl9FBrM5HuzZm6cMmlp40jIoCMHKeH42xtJBTOPzbkG/z9aWaNagjn8jEFKWB28w4hjufN6NG1QReF2ai7befjTnRmg==",
								},
							},
						},
					},
				},
				Subjects: map[digest.Digest]*ocispecs.SubjectDescriptor{
					subjectDigest: {
						Descriptor: imgspec.Descriptor{
							Digest:    subjectDigest,
							MediaType: imgspec.MediaTypeImageManifest,
						},
					},
				},
				Blobs: map[digest.Digest][]byte{
					"sha256:6e1ffef2ba058cda5d1aa7ed792cb1e63b4207d8195a469bee1b5fc662cd9b70": []byte(`{"critical":{"identity":{"docker-reference":"artifactstest.azurecr.io/4-15-24/cosign/hello-world"},"image":{"docker-manifest-digest":"sha256:d37ada95d47ad12224c205a938129df7a3e52345828b4fa27b03a98825d1e2e7"},"type":"cosign container image signature"},"optional":null}`),
				},
			},
			expectedResultMessagePrefix: "cosign verification success",
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			getKeyMapOpts = func(_ context.Context, _ TrustPolicy, _ string) (map[PKKey]keymanagementprovider.PublicKey, cosign.CheckOpts, error) {
				if tt.getKeysError {
					return nil, cosign.CheckOpts{}, fmt.Errorf("error")
				}

				return tt.keys, tt.cosignOpts, nil
			}
			verifierFactory := cosignVerifierFactory{}
			trustPoliciesConfig := []TrustPolicyConfig{
				{
					Name:    "test-policy",
					Keyless: KeylessConfig{CertificateIdentity: testIdentity, CertificateOIDCIssuer: testIssuer},
					Scopes:  []string{"*"},
				},
			}
			validConfig := config.VerifierConfig{
				"name":          "test",
				"artifactTypes": "testtype",
				"type":          "cosign",
				"trustPolicies": trustPoliciesConfig,
			}
			cosignVerifier, err := verifierFactory.Create("", validConfig, "", "test-namespace")
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}
			result, _ := cosignVerifier.Verify(context.Background(), subjectRef, refDescriptor, tt.store)
			if !strings.HasPrefix(result.Message, tt.expectedResultMessagePrefix) {
				t.Errorf("Verify() = %v, want %v", result.Message, tt.expectedResultMessagePrefix)
			}
		})
	}
}

// TestVerificationMessage tests the verificationMessage function
func TestVerificationMessage(t *testing.T) {
	tc := []struct {
		name             string
		expectedMessages []string
		bundleVerified   bool
		checkOpts        cosign.CheckOpts
	}{
		{
			name:             "keyed, offline bundle, claims with annotations",
			expectedMessages: []string{annotationMessage, claimsMessage, offlineBundleMessage, sigVerifierMessage},
			bundleVerified:   true,
			checkOpts: cosign.CheckOpts{
				ClaimVerifier: cosign.SimpleClaimVerifier,
				Annotations: map[string]interface{}{
					"test": "test",
				},
				SigVerifier: &mockNoOpVerifier{},
			},
		},
		{
			name:             "keyless, rekor, fulcio",
			expectedMessages: []string{rekorClaimsMessage, rekorSigMessage, certVerifierMessage},
			bundleVerified:   false,
			checkOpts: cosign.CheckOpts{
				RekorClient: &client.Rekor{},
			},
		},
	}
	for i, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			result := verificationPerformedMessage(tt.bundleVerified, &tc[i].checkOpts)
			if !slices.Equal(result, tt.expectedMessages) {
				t.Errorf("verificationMessage() = %v, want %v", result, tt.expectedMessages)
			}
		})
	}
}
