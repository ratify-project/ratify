package notaryv2

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	paths "path/filepath"
	"reflect"
	"testing"
	"time"

	ratifyconfig "github.com/deislabs/ratify/config"
	"github.com/deislabs/ratify/pkg/common"
	e "github.com/deislabs/ratify/pkg/executor"
	"github.com/deislabs/ratify/pkg/executor/types"
	"github.com/deislabs/ratify/pkg/homedir"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore"
	"github.com/deislabs/ratify/pkg/referrerstore/config"
	"github.com/deislabs/ratify/pkg/verifier"
	sig "github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/verifier/truststore"
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
	testExecutor                         = &mockExecutor{}
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

type mockExecutor struct{}

func (e mockExecutor) VerifySubject(ctx context.Context, verifyParameters e.VerifyParameters) (types.VerifyResult, error) {
	return types.VerifyResult{}, nil
}

func (e mockExecutor) GetVerifyRequestTimeout() time.Duration {
	return time.Hour
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

			result, err := v.Verify(context.Background(), tt.ref, ocispecs.ReferenceDescriptor{}, store, testExecutor)

			if (err != nil) != tt.expectErr {
				t.Fatalf("error = %v, expectErr = %v", err, tt.expectErr)
			}
			if result.IsSuccess != tt.expect.IsSuccess {
				t.Fatalf("expect %+v, got %+v", tt.expect, result)
			}
		})
	}
}

func TestGetCertificates_EmptyCertMap(t *testing.T) {
	certStore := map[string][]string{}
	certStore["store1"] = []string{"kv1"}
	certStore["store2"] = []string{"kv2"}
	store := &trustStore{
		certStores: certStore,
	}

	certificatesMap := map[string][]*x509.Certificate{}
	_, err := store.getCertificatesInternal(context.Background(), truststore.TypeCA, "store1", certificatesMap)

	if err == nil {
		t.Fatalf("error expected if cert map is empty")
	}
}

func TestGetCertificates_NamedStore(t *testing.T) {
	certStore := map[string][]string{}
	certStore["store1"] = []string{"kv1"}
	certStore["store2"] = []string{"kv2"}

	store := &trustStore{
		certStores: certStore,
	}

	kv1Cert := getCert("-----BEGIN CERTIFICATE-----\nMIIDsDCCApigAwIBAgIQMdNmNTKwQ9aOe6iuMRokDzANBgkqhkiG9w0BAQsFADBa\nMQswCQYDVQQGEwJVUzELMAkGA1UECBMCV0ExEDAOBgNVBAcTB1NlYXR0bGUxDzAN\nBgNVBAoTBk5vdGFyeTEbMBkGA1UEAxMSd2FiYml0LW5ldHdvcmtzLmlvMB4XDTIy\nMTIxNDIxNTAzMVoXDTIzMTIxNDIyMDAzMVowWjELMAkGA1UEBhMCVVMxCzAJBgNV\nBAgTAldBMRAwDgYDVQQHEwdTZWF0dGxlMQ8wDQYDVQQKEwZOb3RhcnkxGzAZBgNV\nBAMTEndhYmJpdC1uZXR3b3Jrcy5pbzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCC\nAQoCggEBAOP6AHCFz41kRqsAiv6guFtQVsrzMgzoCX7o9NtQ57rr8BESP1LTGRAO\n4bjyP0i+at5uwIm4tdz0gW+g0P+f9bmfiScYgOFuxAJxLkMkBWPN3dJ9ulP/OGgB\n6mSCsEGreB3uaGc5rMbWCRaux65bMPjEzx5ex0qRSsn+fFMTwINPQUJpXSvi/W2/\n1umEWE1x59x0vlkP2dN7CXtB5/Bh01QNNbMdKU9saYn0kaBrCYZLwr6AxFRzLqLM\nQggy/6bOp/+cTTVqTiChMcdyIX52GRr2lChRsB34dDPYxDeKSI5LoRy07bveLjex\n4wm9+vx/WOSS5z0QPvE/v8avuIkMXR0CAwEAAaNyMHAwDgYDVR0PAQH/BAQDAgeA\nMAkGA1UdEwQCMAAwEwYDVR0lBAwwCgYIKwYBBQUHAwMwHwYDVR0jBBgwFoAUwVvE\nvqQPxnE6j6pfX6jpSyv2dOAwHQYDVR0OBBYEFMFbxL6kD8ZxOo+qX1+o6Usr9nTg\nMA0GCSqGSIb3DQEBCwUAA4IBAQDE61FLbagvlCcXf0zcv+mUQ+0HvDVs7ofQe3Yw\naz7gAwxgTspr+jIFQWnPOOBupsyx/jucoz78ndbc5DGWPs2Qz/pIEGnLto2W/PYy\nas/9n8xHxembS4n/Mxxp60PF6ladi/nJAtDJds67sBeqLOfJzh6jV2uQvW7PXe1P\nOMSUHbBn8AfArZ/9njusiLs75+XcAgpnBFqKVv2Vd/INp2YQpVzusuiodeM8A9Qt\n/5xykjdCJw3ceZxD7dSkHgchKZPINFBYHt/EkN/d8mXFOKjGXZyntp4PO6PJ2HYN\nhMMDwdNu4mBmlMTdZMPEpIZIeW7G0P9KpCuvvD7po7NxdBgI\n-----END CERTIFICATE-----\n")
	kv2Cert := getCert("-----BEGIN CERTIFICATE-----\nMIIDsDCCApigAwIBAgIQFJMQeqR8TRuHqNu+x0MuEDANBgkqhkiG9w0BAQsFADBa\nMQswCQYDVQQGEwJVUzELMAkGA1UECBMCV0ExEDAOBgNVBAcTB1NlYXR0bGUxDzAN\nBgNVBAoTBk5vdGFyeTEbMBkGA1UEAxMSd2FiYml0LW5ldHdvcmtzLmlvMB4XDTIz\nMDExMTE5MjAxMloXDTI0MDExMTE5MzAxMlowWjELMAkGA1UEBhMCVVMxCzAJBgNV\nBAgTAldBMRAwDgYDVQQHEwdTZWF0dGxlMQ8wDQYDVQQKEwZOb3RhcnkxGzAZBgNV\nBAMTEndhYmJpdC1uZXR3b3Jrcy5pbzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCC\nAQoCggEBAMh7F6sZyeiQRva83SvQu0PbsyD4zkEeWAyj03n1dx91FEeEXItCr+Y1\nghQKgdBOY/wJQmSq/We1e+17NoNICrzy2Y1sOVMYR5sx8H/UxO3q8oS7bxctFy+e\nHs4BxlHIqeIiWnz9bFAJFqV6BkJDVjp9k5QfHlkqH08WBvm/D8YTpWzvEPn+71ZG\nN1RKqFUeeM949oGGnC63vVMRRYIx2LoJliNZXdj9qoOHZksDrX2jkgPykkOYcmfo\n9CH9v0JNX+0t0Enp0ruUFK1pSZW+TicI22GvENYHGZNZ0m+6oD5ePRZoYhWyAzgZ\nndHO5bYh3yC7DMc6ssOEJeNN0I2+iLUCAwEAAaNyMHAwDgYDVR0PAQH/BAQDAgeA\nMAkGA1UdEwQCMAAwEwYDVR0lBAwwCgYIKwYBBQUHAwMwHwYDVR0jBBgwFoAUYhhf\nPFgAqU8PF3ClvfKs67HmpWwwHQYDVR0OBBYEFGIYXzxYAKlPDxdwpb3yrOux5qVs\nMA0GCSqGSIb3DQEBCwUAA4IBAQCXu1w+6s2RO2/KPmC+29m9EjbDReI4bGlDGgiv\nwk1fmvPvDrqL4Ebpcrb1nstNlsxpKYQP+3Vi8gPiqNQ7JvPStd1NBu+ViCXdvOe5\nCtN7tBFTCBgdgXNZ9bvIM2dS+xW/ZAJdyHbV9Hn77+rs/uCDHtbaQMJ3N9LGW8GR\nGY+uYylrrCrjb9fzotMaONnF9c1GKiANskc9371wbaninpxcwMNA5j027XzfnMEW\nm807wjlNV3Kuf4fdDpzBLe940iplfTlQMylWMqgANpEw4EqHCrBJPQAHfQEpQlo+\n9H72lrqOiYNNwApfB9P+UqMDi1B7T2yzfvXcqQ75FpSRIxzK\n-----END CERTIFICATE-----\n")

	certificatesMap := map[string][]*x509.Certificate{}
	certificatesMap["kv1"] = []*x509.Certificate{kv1Cert}
	certificatesMap["kv2"] = []*x509.Certificate{kv2Cert}

	// only the certificate in the specified namedStore should be returned
	result, _ := store.getCertificatesInternal(context.Background(), truststore.TypeCA, "store1", certificatesMap)
	expectedLen := 1

	if len(result) != expectedLen {
		t.Fatalf("unexpected count of certificate, expected %+v, got %+v", expectedLen, len(result))
	}

	if !reflect.DeepEqual(result[0], kv1Cert) {
		t.Fatalf("unexpected certificate returned")
	}
}

// convert string to a x509 certificate
func getCert(certString string) *x509.Certificate {
	block, _ := pem.Decode([]byte(certString))
	if block == nil {
		panic("failed to parse certificate PEM")
	}

	test, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic("failed to parse certificate: " + err.Error())
	}

	return test
}
