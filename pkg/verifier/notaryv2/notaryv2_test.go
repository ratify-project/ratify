package notaryv2

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
	testNotaryVerifier notation.Verifier = mockNotaryVerifier{}
	validBlobDesc                        = ocispec.Descriptor{
		Digest: testDigest,
	}
	validBlobDesc2 = ocispec.Descriptor{
		Digest: testDigest2,
	}
	invalidBlobDesc = ocispec.Descriptor{
		Digest: invalidDigest,
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

type mockNotaryVerifier struct{}

func (v mockNotaryVerifier) Verify(ctx context.Context, desc ocispec.Descriptor, signature []byte, opts notation.VerifyOptions) (*notation.VerificationOutcome, error) {
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

func (s mockStore) ListReferrers(ctx context.Context, subjectReference common.Reference, artifactTypes []string, nextToken string, subjectDesc *ocispecs.SubjectDescriptor) (referrerstore.ListReferrersResult, error) {
	return referrerstore.ListReferrersResult{}, nil
}

func (s mockStore) GetBlobContent(ctx context.Context, subjectReference common.Reference, digest digest.Digest) ([]byte, error) {
	if s.refBlob == nil {
		return nil, fmt.Errorf("invalid blob")
	}
	return s.refBlob, nil
}

func (s mockStore) GetReferenceManifest(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor) (ocispecs.ReferenceManifest, error) {
	if len(s.manifest.Blobs) == 0 {
		return s.manifest, fmt.Errorf("invalid reference")
	}
	return s.manifest, nil
}

func (s mockStore) GetConfig() *config.StoreConfig {
	return nil
}

func (s mockStore) GetSubjectDescriptor(ctx context.Context, subjectReference common.Reference) (*ocispecs.SubjectDescriptor, error) {
	return &ocispecs.SubjectDescriptor{
		Descriptor: ocispec.Descriptor{},
	}, nil
}

func TestName(t *testing.T) {
	v := &notaryV2Verifier{}
	name := v.Name()

	if name != "notaryv2" {
		t.Fatalf("expect name: notaryv2, got: %s", name)
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
			v := &notaryV2Verifier{
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
		expect    *NotaryV2VerifierConfig
	}{
		{
			name: "failed unmarshalling to notary config",
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
			expect: &NotaryV2VerifierConfig{
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
			expect: &NotaryV2VerifierConfig{
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
					"certstore1": {"akv1", "akv2"},
					"certstore2": {"akv3", "akv4"},
				},
			},
			expectErr: false,
			expect: &NotaryV2VerifierConfig{
				Name:              test,
				VerificationCerts: []string{testPath, defaultCertDir},
				VerificationCertStores: map[string][]string{
					"certstore1": {"akv1", "akv2"},
					"certstore2": {"akv3", "akv4"},
				},
			},
		},
	}

	//TODO add new test for parseVerifierConfig
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notaryConfig, err := parseVerifierConfig(tt.configMap)

			if (err != nil) != tt.expectErr {
				t.Errorf("error = %v, expectErr = %v", err, tt.expectErr)
			}
			if !reflect.DeepEqual(notaryConfig, tt.expect) {
				t.Errorf("expect %+v, got %+v", tt.expect, notaryConfig)
			}
		})
	}
}

func TestVerifySignature(t *testing.T) {
	v := &notaryV2Verifier{
		notationVerifier: &testNotaryVerifier,
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
			f := &notaryv2VerifierFactory{}
			_, err := f.Create(testVersion, tt.configMap)

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
			v := &notaryV2Verifier{
				notationVerifier: &testNotaryVerifier,
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
	verifier := &notaryV2Verifier{}
	nestedReferences := verifier.GetNestedReferences()

	if len(nestedReferences) != 0 {
		t.Fatalf("notation signature should not have nested references")
	}
}
