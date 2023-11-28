// Copyright The Ratify Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package notation

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	paths "path/filepath"
	"reflect"
	"testing"

	ratifyconfig "github.com/deislabs/ratify/config"
	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/homedir"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore"
	"github.com/deislabs/ratify/pkg/referrerstore/config"
	"github.com/deislabs/ratify/pkg/verifier"
	sig "github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation-go"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	testArtifactType1 = "testArtifactType1"
	testArtifactType2 = "testArtifactType2"
	test              = "test"
	testPath          = "/test/path"
	testMediaType     = "testMediaType"
	testVersion       = "1.0"
	testDigest        = "sha256:123456"
	testDigest2       = "sha256:234567"
	invalidDigest     = "invalidDigest"
)

var (
	failedResult = verifier.VerifierResult{IsSuccess: false}
	testOutcome  = &notation.VerificationOutcome{
		EnvelopeContent: &sig.EnvelopeContent{
			SignerInfo: sig.SignerInfo{
				CertificateChain: []*x509.Certificate{
					{Issuer: pkix.Name{}, Subject: pkix.Name{}},
				},
			},
		},
	}
	testRefBlob  = []byte("test")
	testRefBlob2 = []byte("test2")
	testDesc1    = ocispec.Descriptor{}
	validRef     = common.Reference{
		Original: "testRegistry/repo:v1",
		Digest:   testDigest,
	}
	validRef2 = common.Reference{
		Original: "testRegistry/repo:v2",
		Digest:   testDigest2,
	}
	invalidRef = common.Reference{
		Original: "invalid",
	}
	testNotationPluginVerifier notation.Verifier = mockNotationPluginVerifier{}
	validBlobDesc                                = ocispec.Descriptor{
		Digest: testDigest,
	}
	validBlobDesc2 = ocispec.Descriptor{
		Digest: testDigest2,
	}
	defaultCertDir  = paths.Join(homedir.Get(), ratifyconfig.ConfigFileDir, defaultCertPath)
	testTrustPolicy = map[string]interface{}{
		"version": "1.0",
		"trustPolicies": []map[string]interface{}{
			{
				"name":           "default",
				"registryScopes": []string{"*"},
				"signatureVerification": map[string]string{
					"level": "strict",
				},
				"trustStores":       []string{"ca:certs"},
				"trustedIdentities": []string{"*"},
			},
		},
	}
)

type mockNotationPluginVerifier struct{}

func (v mockNotationPluginVerifier) Verify(_ context.Context, _ ocispec.Descriptor, signature []byte, _ notation.VerifierVerifyOptions) (*notation.VerificationOutcome, error) {
	if reflect.DeepEqual(signature, testRefBlob2) {
		return nil, fmt.Errorf("failed verification")
	}
	return testOutcome, nil
}

type mockStore struct {
	refBlob  []byte
	manifest ocispecs.ReferenceManifest
}

func (s mockStore) Name() string {
	return test
}

func (s mockStore) ListReferrers(_ context.Context, _ common.Reference, _ []string, _ string, _ *ocispecs.SubjectDescriptor) (referrerstore.ListReferrersResult, error) {
	return referrerstore.ListReferrersResult{}, nil
}

func (s mockStore) GetBlobContent(_ context.Context, _ common.Reference, _ digest.Digest) ([]byte, error) {
	if s.refBlob == nil {
		return nil, fmt.Errorf("invalid blob")
	}
	return s.refBlob, nil
}

func (s mockStore) GetReferenceManifest(_ context.Context, _ common.Reference, _ ocispecs.ReferenceDescriptor) (ocispecs.ReferenceManifest, error) {
	if len(s.manifest.Blobs) == 0 {
		return s.manifest, fmt.Errorf("invalid reference")
	}
	return s.manifest, nil
}

func (s mockStore) GetConfig() *config.StoreConfig {
	return nil
}

func (s mockStore) GetSubjectDescriptor(_ context.Context, _ common.Reference) (*ocispecs.SubjectDescriptor, error) {
	return &ocispecs.SubjectDescriptor{
		Descriptor: ocispec.Descriptor{},
	}, nil
}

func TestName(t *testing.T) {
	v := &notationPluginVerifier{}
	name := v.Name()

	if name != "notation" {
		t.Fatalf("expect name: notation, got: %s", name)
	}
}

func TestCanVerify(t *testing.T) {
	tests := []struct {
		name              string
		artifactTypes     []string
		referenceArtifact string
		expect            bool
	}{
		{
			name:              "wildcard pattern",
			artifactTypes:     []string{"*"},
			referenceArtifact: testArtifactType1,
			expect:            true,
		},
		{
			name:              "type unmatched",
			artifactTypes:     []string{testArtifactType2},
			referenceArtifact: testArtifactType1,
			expect:            false,
		},
		{
			name:              "type matched",
			artifactTypes:     []string{testArtifactType1},
			referenceArtifact: testArtifactType1,
			expect:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &notationPluginVerifier{
				artifactTypes: tt.artifactTypes,
			}
			desc := ocispecs.ReferenceDescriptor{
				ArtifactType: tt.referenceArtifact,
			}

			got := v.CanVerify(context.Background(), desc)
			if got != tt.expect {
				t.Fatalf("Expect: %v, got: %v", tt.expect, got)
			}
		})
	}
}

func TestParseVerifierConfig(t *testing.T) {
	tests := []struct {
		name      string
		configMap map[string]interface{}
		expectErr bool
		expect    *NotationPluginVerifierConfig
	}{
		{
			name: "failed unmarshalling to notation config",
			configMap: map[string]interface{}{
				"name": []string{test},
			},
			expectErr: true,
			expect:    nil,
		},
		{
			name: "successfully parsed with default cert directory",
			configMap: map[string]interface{}{
				"name": test,
			},
			expectErr: false,
			expect: &NotationPluginVerifierConfig{
				Name:              test,
				VerificationCerts: []string{defaultCertDir},
			},
		},
		{
			name: "successfully parsed with specified cert directory",
			configMap: map[string]interface{}{
				"name":              test,
				"verificationCerts": []string{testPath},
			},
			expectErr: false,
			expect: &NotationPluginVerifierConfig{
				Name:              test,
				VerificationCerts: []string{testPath, defaultCertDir},
			},
		},
		{
			name: "successfully parsed with specified cert stores",
			configMap: map[string]interface{}{
				"name":              test,
				"verificationCerts": []string{testPath},
				"verificationCertStores": map[string][]string{
					"certstore1": {"defaultns/akv1", "akv2"},
					"certstore2": {"akv3", "akv4"},
				},
			},
			expectErr: false,
			expect: &NotationPluginVerifierConfig{
				Name:              test,
				VerificationCerts: []string{testPath, defaultCertDir},
				VerificationCertStores: map[string][]string{
					"certstore1": {"defaultns/akv1", "testns/akv2"},
					"certstore2": {"testns/akv3", "testns/akv4"},
				},
			},
		},
	}

	//TODO add new test for parseVerifierConfig
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notationPluginConfig, err := parseVerifierConfig(tt.configMap, "testns")

			if (err != nil) != tt.expectErr {
				t.Errorf("error = %v, expectErr = %v", err, tt.expectErr)
			}
			if !reflect.DeepEqual(notationPluginConfig, tt.expect) {
				t.Errorf("expect %+v, got %+v", tt.expect, notationPluginConfig)
			}
		})
	}
}

func TestVerifySignature(t *testing.T) {
	v := &notationPluginVerifier{
		notationVerifier: &testNotationPluginVerifier,
	}

	outcome, err := v.verifySignature(context.Background(), testArtifactType1, testMediaType, testDesc1, testRefBlob)
	if err != nil {
		t.Errorf("got unexpected error: %v", err)
	}
	if !reflect.DeepEqual(outcome, testOutcome) {
		t.Fatalf("expect outcome: %v, got: %v", outcome, testOutcome)
	}
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name      string
		configMap map[string]interface{}
		expect    verifier.ReferenceVerifier
		expectErr bool
	}{
		{
			name: "failed parsing verifier config",
			configMap: map[string]interface{}{
				"name": []string{test},
			},
			expectErr: true,
		},
		{
			name: "failed loading policy",
			configMap: map[string]interface{}{
				"name":        test,
				"trustPolicy": "{",
			},
			expectErr: true,
		},
		{
			name: "created verifier successfully",
			configMap: map[string]interface{}{
				"name":           test,
				"trustPolicyDoc": testTrustPolicy,
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &notationPluginVerifierFactory{}
			_, err := f.Create(testVersion, tt.configMap, "", "")

			if (err != nil) != tt.expectErr {
				t.Fatalf("error = %v, expectErr = %v", err, tt.expectErr)
			}
		})
	}
}

func TestVerify(t *testing.T) {
	tests := []struct {
		name      string
		expect    verifier.VerifierResult
		ref       common.Reference
		manifest  ocispecs.ReferenceManifest
		refBlob   []byte
		expectErr bool
	}{
		{
			name:      "failed getting manifest",
			ref:       invalidRef,
			refBlob:   []byte(""),
			manifest:  ocispecs.ReferenceManifest{},
			expect:    failedResult,
			expectErr: true,
		},
		{
			name:    "failed verifying signature",
			ref:     validRef2,
			refBlob: testRefBlob2,
			manifest: ocispecs.ReferenceManifest{
				Blobs: []ocispec.Descriptor{validBlobDesc2},
			},
			expect:    failedResult,
			expectErr: true,
		},
		{
			name:    "verified successfully",
			ref:     validRef,
			refBlob: testRefBlob,
			manifest: ocispecs.ReferenceManifest{
				Blobs: []ocispec.Descriptor{validBlobDesc},
			},
			expect:    verifier.VerifierResult{IsSuccess: true},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &notationPluginVerifier{
				notationVerifier: &testNotationPluginVerifier,
			}

			store := &mockStore{
				refBlob:  tt.refBlob,
				manifest: tt.manifest,
			}

			result, err := v.Verify(context.Background(), tt.ref, ocispecs.ReferenceDescriptor{}, store)

			if (err != nil) != tt.expectErr {
				t.Fatalf("error = %v, expectErr = %v", err, tt.expectErr)
			}
			if result.IsSuccess != tt.expect.IsSuccess {
				t.Fatalf("expect %+v, got %+v", tt.expect, result)
			}
		})
	}
}

func TestGetNestedReferences(t *testing.T) {
	verifier := &notationPluginVerifier{}
	nestedReferences := verifier.GetNestedReferences()

	if len(nestedReferences) != 0 {
		t.Fatalf("notation signature should not have nested references")
	}
}
