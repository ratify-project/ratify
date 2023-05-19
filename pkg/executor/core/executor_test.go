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

package core

// import (
// 	"context"
// 	"errors"
// 	"reflect"
// 	"testing"
// 	"time"

// 	"github.com/deislabs/ratify/pkg/common"
// 	exConfig "github.com/deislabs/ratify/pkg/executor/config"
// 	"github.com/deislabs/ratify/pkg/policyprovider"

// 	e "github.com/deislabs/ratify/pkg/executor"
// 	"github.com/deislabs/ratify/pkg/executor/config"
// 	"github.com/deislabs/ratify/pkg/executor/types"
// 	"github.com/deislabs/ratify/pkg/ocispecs"
// 	policyConfig "github.com/deislabs/ratify/pkg/policyprovider/configpolicy"
// 	policyTypes "github.com/deislabs/ratify/pkg/policyprovider/types"
// 	"github.com/deislabs/ratify/pkg/referrerstore"
// 	storeConfig "github.com/deislabs/ratify/pkg/referrerstore/config"
// 	"github.com/deislabs/ratify/pkg/referrerstore/mocks"
// 	"github.com/deislabs/ratify/pkg/verifier"
// 	"github.com/opencontainers/go-digest"
// 	oci "github.com/opencontainers/image-spec/specs-go/v1"
// )

// const (
// 	testArtifactType1 = "test-type1"
// 	testArtifactType2 = "test-type2"
// 	subject1          = "localhost:5000/net-monitor:v1"
// 	subjectDigest     = "sha256:6a5a5368e0c2d3e5909184fa28ddfd56072e7ff3ee9a945876f7eee5896ef5bb"
// 	signatureDigest   = "sha256:9f13e0ac480cf86a5c9ec5d173001bbb6ec455f501f1812f0b0ad1f3468e8cfa"
// 	artifactType      = "testArtifactType"
// )

// type mockPolicyProvider struct {
// 	result bool
// }

// func (p *mockPolicyProvider) VerifyNeeded(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor) bool {
// 	return true
// }

// func (p *mockPolicyProvider) ContinueVerifyOnFailure(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor, partialVerifyResult types.VerifyResult) bool {
// 	return true
// }

// func (p *mockPolicyProvider) ErrorToVerifyResult(ctx context.Context, subjectRefString string, verifyError error) types.VerifyResult {
// 	return types.VerifyResult{}
// }

// func (p *mockPolicyProvider) OverallVerifyResult(ctx context.Context, verifierReports []interface{}) bool {
// 	return p.result
// }

// type mockStore struct {
// 	referrers map[string][]ocispecs.ReferenceDescriptor
// }

// func (s *mockStore) Name() string {
// 	return "mockStore"
// }

// func (s *mockStore) ListReferrers(ctx context.Context, subjectReference common.Reference, artifactTypes []string, nextToken string, subjectDesc *ocispecs.SubjectDescriptor) (referrerstore.ListReferrersResult, error) {
// 	if s.referrers == nil {
// 		return referrerstore.ListReferrersResult{}, errors.New("some error happened")
// 	}
// 	if _, ok := s.referrers[subjectDesc.Digest.String()]; ok {
// 		return referrerstore.ListReferrersResult{
// 			NextToken: "",
// 			Referrers: s.referrers[subjectDesc.Digest.String()],
// 		}, nil
// 	}
// 	return referrerstore.ListReferrersResult{}, nil
// }

// func (s *mockStore) GetBlobContent(ctx context.Context, subjectReference common.Reference, digest digest.Digest) ([]byte, error) {
// 	return nil, nil
// }

// func (s *mockStore) GetReferenceManifest(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor) (ocispecs.ReferenceManifest, error) {
// 	return ocispecs.ReferenceManifest{}, nil
// }

// func (s *mockStore) GetConfig() *storeConfig.StoreConfig {
// 	return nil
// }

// func (s *mockStore) GetSubjectDescriptor(ctx context.Context, subjectReference common.Reference) (*ocispecs.SubjectDescriptor, error) {
// 	if subjectReference.Tag == "v1" {
// 		return &ocispecs.SubjectDescriptor{
// 			Descriptor: oci.Descriptor{
// 				Digest: subjectDigest,
// 			},
// 		}, nil
// 	}
// 	return &ocispecs.SubjectDescriptor{
// 		Descriptor: oci.Descriptor{
// 			Digest: subjectReference.Digest,
// 		},
// 	}, nil
// }

// type mockVerifier struct {
// 	canVerify      bool
// 	verifierResult verifier.VerifierResult
// }

// func (v *mockVerifier) Name() string {
// 	return "mockVerifier"
// }

// func (v *mockVerifier) CanVerify(ctx context.Context, referenceDescriptor ocispecs.ReferenceDescriptor) bool {
// 	return v.canVerify
// }

// func (v *mockVerifier) Verify(ctx context.Context,
// 	subjectReference common.Reference,
// 	referenceDescriptor ocispecs.ReferenceDescriptor,
// 	referrerStore referrerstore.ReferrerStore) (verifier.VerifierResult, error) {
// 	if reflect.DeepEqual(v.verifierResult, verifier.VerifierResult{}) {
// 		return verifier.VerifierResult{}, errors.New("no verifier result")
// 	}
// 	return v.verifierResult, nil
// }

// func (v *mockVerifier) GetNestedReferences() []string {
// 	return nil
// }

// func TestNewExecutor(t *testing.T) {
// 	testCases := []struct {
// 		name      string
// 		config    *config.ExecutorConfig
// 		expectErr bool
// 	}{
// 		{
// 			name: "invalid config",
// 			config: &config.ExecutorConfig{
// 				ExecutionMode: config.PassthroughExecutionMode,
// 				UseRegoPolicy: false,
// 			},
// 			expectErr: true,
// 		},
// 		{
// 			name:      "valid config",
// 			config:    &config.ExecutorConfig{},
// 			expectErr: false,
// 		},
// 		{
// 			name:      "nil config",
// 			config:    nil,
// 			expectErr: false,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			_, err := NewExecutor(nil, nil, nil, tc.config)

// 			if (err != nil) != tc.expectErr {
// 				t.Fatalf("expected error: %v, got: %v", tc.expectErr, err)
// 			}
// 		})
// 	}
// }

// func TestVerifySubjectInternalWithDecision_ResolveSubjectDescriptor_Failed(t *testing.T) {
// 	executor := Executor{}

// 	verifyParameters := e.VerifyParameters{
// 		Subject: "localhost:5000/net-monitor:v1",
// 	}

// 	_, err := executor.verifySubjectInternalWithDecision(context.Background(), verifyParameters)

// 	if err == nil {
// 		t.Fatal("expected subject parsing to fail")
// 	}
// }

// func TestVerifySubjectInternalWithDecision_ResolveSubjectDescriptor_Success(t *testing.T) {
// 	testDigest := digest.FromString("test")
// 	store := &mocks.TestStore{
// 		References: []ocispecs.ReferenceDescriptor{},
// 		ResolveMap: map[string]digest.Digest{
// 			"v1": testDigest,
// 		},
// 	}

// 	executor := Executor{
// 		ReferrerStores: []referrerstore.ReferrerStore{store},
// 	}

// 	verifyParameters := e.VerifyParameters{
// 		Subject: "localhost:5000/net-monitor:v1",
// 	}

// 	if _, err := executor.verifySubjectInternalWithDecision(context.Background(), verifyParameters); !errors.Is(err, ErrReferrersNotFound) {
// 		t.Fatalf("expected ErrReferrersNotFound actual %v", err)
// 	}
// }

// func TestVerifySubjectInternalWithDecision_Verify_NoReferrers(t *testing.T) {
// 	testDigest := digest.FromString("test")
// 	configPolicy := policyConfig.PolicyEnforcer{}
// 	ex := &Executor{
// 		PolicyEnforcer: configPolicy,
// 		ReferrerStores: []referrerstore.ReferrerStore{&mocks.TestStore{
// 			ResolveMap: map[string]digest.Digest{
// 				"v1": testDigest,
// 			},
// 		}},
// 		Verifiers: []verifier.ReferenceVerifier{&TestVerifier{}},
// 		Config: &exConfig.ExecutorConfig{
// 			VerificationRequestTimeout: nil,
// 			MutationRequestTimeout:     nil,
// 		},
// 	}

// 	verifyParameters := e.VerifyParameters{
// 		Subject: "localhost:5000/net-monitor:v1",
// 	}

// 	if _, err := ex.verifySubjectInternalWithDecision(context.Background(), verifyParameters); !errors.Is(err, ErrReferrersNotFound) {
// 		t.Fatalf("expected ErrReferrersNotFound actual %v", err)
// 	}
// }

// func TestVerifySubjectInternalWithDecision_CanVerify_ExpectedResults(t *testing.T) {
// 	testDigest := digest.FromString("test")
// 	configPolicy := policyConfig.PolicyEnforcer{
// 		ArtifactTypePolicies: map[string]policyTypes.ArtifactTypeVerifyPolicy{
// 			testArtifactType1: policyTypes.AnyVerifySuccess,
// 		}}
// 	store := &mocks.TestStore{References: []ocispecs.ReferenceDescriptor{
// 		{
// 			ArtifactType: testArtifactType1,
// 		},
// 		{
// 			ArtifactType: "test-type2",
// 		}},
// 		ResolveMap: map[string]digest.Digest{
// 			"v1": testDigest,
// 		},
// 	}
// 	ver := &TestVerifier{
// 		CanVerifyFunc: func(at string) bool {
// 			return at == testArtifactType1
// 		},
// 		VerifyResult: func(artifactType string) bool {
// 			return true
// 		},
// 	}

// 	ex := &Executor{
// 		PolicyEnforcer: configPolicy,
// 		ReferrerStores: []referrerstore.ReferrerStore{store},
// 		Verifiers:      []verifier.ReferenceVerifier{ver},
// 		Config: &exConfig.ExecutorConfig{
// 			VerificationRequestTimeout: nil,
// 			MutationRequestTimeout:     nil,
// 		},
// 	}

// 	verifyParameters := e.VerifyParameters{
// 		Subject: "localhost:5000/net-monitor:v1",
// 	}

// 	result, err := ex.verifySubjectInternalWithDecision(context.Background(), verifyParameters)

// 	if err != nil {
// 		t.Fatalf("verification failed with err %v", err)
// 	}

// 	if !result.IsSuccess {
// 		t.Fatal("verification expected to be success")
// 	}

// 	if len(result.VerifierReports) != 1 {
// 		t.Fatalf("verification expected to return single report but actual count %d", len(result.VerifierReports))
// 	}
// }

// func TestVerifySubjectInternalWithDecision_VerifyFailures_ExpectedResults(t *testing.T) {
// 	testDigest := digest.FromString("test")
// 	testArtifactType := "test-type1"
// 	configPolicy := policyConfig.PolicyEnforcer{
// 		ArtifactTypePolicies: map[string]policyTypes.ArtifactTypeVerifyPolicy{
// 			testArtifactType: policyTypes.AnyVerifySuccess,
// 		}}
// 	store := &mocks.TestStore{References: []ocispecs.ReferenceDescriptor{
// 		{
// 			ArtifactType: testArtifactType,
// 		},
// 		{
// 			ArtifactType: "test-type2",
// 		}},
// 		ResolveMap: map[string]digest.Digest{
// 			"v1": testDigest,
// 		},
// 	}
// 	ver := &TestVerifier{
// 		CanVerifyFunc: func(at string) bool {
// 			return true
// 		},
// 		VerifyResult: func(artifactType string) bool {
// 			return artifactType != testArtifactType
// 		},
// 	}

// 	ex := &Executor{
// 		PolicyEnforcer: configPolicy,
// 		ReferrerStores: []referrerstore.ReferrerStore{store},
// 		Verifiers:      []verifier.ReferenceVerifier{ver},
// 		Config: &exConfig.ExecutorConfig{
// 			VerificationRequestTimeout: nil,
// 			MutationRequestTimeout:     nil,
// 		},
// 	}

// 	verifyParameters := e.VerifyParameters{
// 		Subject: "localhost:5000/net-monitor:v1",
// 	}

// 	result, err := ex.verifySubjectInternalWithDecision(context.Background(), verifyParameters)

// 	if err != nil {
// 		t.Fatalf("verification failed with err %v", err)
// 	}

// 	if result.IsSuccess {
// 		t.Fatal("verification expected to fail")
// 	}
// }

// func TestVerifySubjectInternalWithDecision_VerifySuccess_ExpectedResults(t *testing.T) {
// 	testDigest := digest.FromString("test")
// 	configPolicy := policyConfig.PolicyEnforcer{
// 		ArtifactTypePolicies: map[string]policyTypes.ArtifactTypeVerifyPolicy{
// 			testArtifactType1: policyTypes.AnyVerifySuccess,
// 			testArtifactType2: policyTypes.AnyVerifySuccess,
// 		}}
// 	store := &mocks.TestStore{References: []ocispecs.ReferenceDescriptor{
// 		{
// 			ArtifactType: testArtifactType1,
// 		},
// 		{
// 			ArtifactType: testArtifactType2,
// 		}},
// 		ResolveMap: map[string]digest.Digest{
// 			"v1": testDigest,
// 		},
// 	}
// 	ver := &TestVerifier{
// 		CanVerifyFunc: func(at string) bool {
// 			return true
// 		},
// 		VerifyResult: func(artifactType string) bool {
// 			return true
// 		},
// 	}

// 	ex := &Executor{
// 		PolicyEnforcer: configPolicy,
// 		ReferrerStores: []referrerstore.ReferrerStore{store},
// 		Verifiers:      []verifier.ReferenceVerifier{ver},
// 		Config: &exConfig.ExecutorConfig{
// 			VerificationRequestTimeout: nil,
// 			MutationRequestTimeout:     nil,
// 		},
// 	}

// 	verifyParameters := e.VerifyParameters{
// 		Subject: "localhost:5000/net-monitor:v1",
// 	}

// 	result, err := ex.verifySubjectInternalWithDecision(context.Background(), verifyParameters)

// 	if err != nil {
// 		t.Fatalf("verification failed with err %v", err)
// 	}

// 	if !result.IsSuccess {
// 		t.Fatal("verification expected to fail")
// 	}

// 	if len(result.VerifierReports) != 2 {
// 		t.Fatalf("verification expected to return two reports but actual count %d", len(result.VerifierReports))
// 	}
// }

// // TestVerifySubjectInternalWithDecision_MultipleArtifacts_ExpectedResults tests multiple artifacts are verified concurrently
// func TestVerifySubjectInternalWithDecision_MultipleArtifacts_ExpectedResults(t *testing.T) {
// 	testDigest := digest.FromString("test")
// 	configPolicy := policyConfig.PolicyEnforcer{
// 		ArtifactTypePolicies: map[string]policyTypes.ArtifactTypeVerifyPolicy{
// 			testArtifactType1: policyTypes.AnyVerifySuccess,
// 			testArtifactType2: policyTypes.AnyVerifySuccess,
// 		}}
// 	store := &mocks.TestStore{References: []ocispecs.ReferenceDescriptor{
// 		{
// 			ArtifactType: testArtifactType1,
// 		},
// 		{
// 			ArtifactType: testArtifactType2,
// 		}},
// 		ResolveMap: map[string]digest.Digest{
// 			"v1": testDigest,
// 		},
// 	}
// 	ver := &TestVerifier{
// 		CanVerifyFunc: func(at string) bool {
// 			return true
// 		},
// 		VerifyResult: func(artifactType string) bool {
// 			if artifactType == testArtifactType1 {
// 				time.Sleep(2 * time.Second)
// 			}
// 			return true
// 		},
// 	}

// 	ex := &Executor{
// 		PolicyEnforcer: configPolicy,
// 		ReferrerStores: []referrerstore.ReferrerStore{store},
// 		Verifiers:      []verifier.ReferenceVerifier{ver},
// 		Config: &exConfig.ExecutorConfig{
// 			VerificationRequestTimeout: nil,
// 			MutationRequestTimeout:     nil,
// 		},
// 	}

// 	verifyParameters := e.VerifyParameters{
// 		Subject: "localhost:5000/net-monitor:v1",
// 	}

// 	result, err := ex.verifySubjectInternalWithDecision(context.Background(), verifyParameters)

// 	if err != nil {
// 		t.Fatalf("verification failed with err %v", err)
// 	}

// 	if !result.IsSuccess {
// 		t.Fatal("verification expected to fail")
// 	}

// 	if len(result.VerifierReports) != 2 {
// 		t.Fatalf("verification expected to return two reports but actual count %d", len(result.VerifierReports))
// 	}

// 	if result.VerifierReports[0].(verifier.VerifierResult).ArtifactType != "test-type2" {
// 		t.Fatalf("verification expected to return second artifact verifier report first")
// 	}
// }

// // TestVerifySubjectInternalWithDecision_NestedReferences_Expected tests verifier config can specify nested references
// func TestVerifySubjectInternalWithDecision_NestedReferences_Expected(t *testing.T) {
// 	configPolicy := policyConfig.PolicyEnforcer{
// 		ArtifactTypePolicies: map[string]policyTypes.ArtifactTypeVerifyPolicy{
// 			"default": "all",
// 		}}

// 	store := mocks.CreateNewTestStoreForNestedSbom()

// 	// sbom verifier WITH nested references in config
// 	sbomVerifier := &TestVerifier{
// 		CanVerifyFunc: func(at string) bool {
// 			return at == mocks.SbomArtifactType
// 		},
// 		VerifyResult: func(artifactType string) bool {
// 			return true
// 		},
// 		nestedReferences: []string{"string-content-does-not-matter"},
// 	}

// 	signatureVerifier := &TestVerifier{
// 		CanVerifyFunc: func(at string) bool {
// 			return at == mocks.SignatureArtifactType
// 		},
// 		VerifyResult: func(artifactType string) bool {
// 			return true
// 		},
// 	}

// 	ex := &Executor{
// 		PolicyEnforcer: configPolicy,
// 		ReferrerStores: []referrerstore.ReferrerStore{store},
// 		Verifiers:      []verifier.ReferenceVerifier{sbomVerifier, signatureVerifier},
// 		Config: &exConfig.ExecutorConfig{
// 			VerificationRequestTimeout: nil,
// 			MutationRequestTimeout:     nil,
// 		},
// 	}

// 	verifyParameters := e.VerifyParameters{
// 		Subject: mocks.TestSubjectWithDigest,
// 	}

// 	result, err := ex.verifySubjectInternalWithDecision(context.Background(), verifyParameters)

// 	if err != nil {
// 		t.Fatalf("verification failed with err %v", err)
// 	}

// 	if !result.IsSuccess {
// 		t.Fatal("verification expected to succeed")
// 	}

// 	if len(result.VerifierReports) != 2 {
// 		t.Fatalf("verification expected to return two reports but actual count %d", len(result.VerifierReports))
// 	}

// 	for _, report := range result.VerifierReports {
// 		castedReport := report.(verifier.VerifierResult)

// 		// check sbom report
// 		if castedReport.ArtifactType == mocks.SbomArtifactType {
// 			// check sbom has one nested results
// 			if len(castedReport.NestedResults) != 1 {
// 				t.Fatalf("Expected sbom report to have 1 nested result")
// 			}
// 			// check sbom nested result is successful
// 			if !castedReport.NestedResults[0].IsSuccess {
// 				t.Fatalf("Expected the sbom nested result to be successful")
// 			}
// 		} else {
// 			// check non-sbom reports have zero nested results
// 			if len(castedReport.NestedResults) != 0 {
// 				t.Fatalf("Expected non-sboms reports to have zero nested results")
// 			}
// 		}
// 	}
// }

// // TestVerifySubjectInternalWithDecision__NoNestedReferences_Expected tests verifier config can specify no nested references
// func TestVerifySubjectInternalWithDecision_NoNestedReferences_Expected(t *testing.T) {
// 	configPolicy := policyConfig.PolicyEnforcer{
// 		ArtifactTypePolicies: map[string]policyTypes.ArtifactTypeVerifyPolicy{
// 			"default": "all",
// 		}}
// 	store := mocks.CreateNewTestStoreForNestedSbom()

// 	// sbom verifier WITHOUT nested references in config
// 	sbomVer := &TestVerifier{
// 		CanVerifyFunc: func(at string) bool {
// 			return at == mocks.SbomArtifactType
// 		},
// 		VerifyResult: func(artifactType string) bool {
// 			return true
// 		},
// 	}

// 	signatureVer := &TestVerifier{
// 		CanVerifyFunc: func(at string) bool {
// 			return at == mocks.SignatureArtifactType
// 		},
// 		VerifyResult: func(artifactType string) bool {
// 			return true
// 		},
// 	}

// 	ex := &Executor{
// 		PolicyEnforcer: configPolicy,
// 		ReferrerStores: []referrerstore.ReferrerStore{store},
// 		Verifiers:      []verifier.ReferenceVerifier{sbomVer, signatureVer},
// 		Config: &exConfig.ExecutorConfig{
// 			VerificationRequestTimeout: nil,
// 			MutationRequestTimeout:     nil,
// 		},
// 	}

// 	verifyParameters := e.VerifyParameters{
// 		Subject: mocks.TestSubjectWithDigest,
// 	}

// 	result, err := ex.verifySubjectInternalWithDecision(context.Background(), verifyParameters)

// 	if err != nil {
// 		t.Fatalf("verification failed with err %v", err)
// 	}

// 	if !result.IsSuccess {
// 		t.Fatal("verification expected to succeed")
// 	}

// 	if len(result.VerifierReports) != 2 {
// 		t.Fatalf("verification expected to return two reports but actual count %d", len(result.VerifierReports))
// 	}

// 	// check each report for: success, zero nested results
// 	for _, report := range result.VerifierReports {
// 		castedReport := report.(verifier.VerifierResult)

// 		// check for success
// 		if !castedReport.IsSuccess {
// 			t.Fatal("verification expected to succeed")
// 		}
// 		// check there are no nested results
// 		if len(castedReport.NestedResults) != 0 {
// 			t.Fatalf("expected reports to have zero nested results")
// 		}
// 	}
// }

// // TestGetVerifyRequestTimeout_ExpectedResults tests the verification request timeout returned
// func TestGetVerifyRequestTimeout_ExpectedResults(t *testing.T) {
// 	testcases := []struct {
// 		setTimeout      int
// 		ex              Executor
// 		expectedTimeout int
// 	}{
// 		{
// 			setTimeout: -1,
// 			ex: Executor{
// 				PolicyEnforcer: policyConfig.PolicyEnforcer{},
// 				ReferrerStores: []referrerstore.ReferrerStore{},
// 				Verifiers:      []verifier.ReferenceVerifier{},
// 				Config:         nil,
// 			},
// 			expectedTimeout: 2900,
// 		},
// 		{
// 			setTimeout: -1,
// 			ex: Executor{
// 				PolicyEnforcer: policyConfig.PolicyEnforcer{},
// 				ReferrerStores: []referrerstore.ReferrerStore{},
// 				Verifiers:      []verifier.ReferenceVerifier{},
// 				Config: &exConfig.ExecutorConfig{
// 					VerificationRequestTimeout: nil,
// 					MutationRequestTimeout:     nil,
// 				},
// 			},
// 			expectedTimeout: 2900,
// 		},
// 		{
// 			setTimeout: 5000,
// 			ex: Executor{
// 				PolicyEnforcer: policyConfig.PolicyEnforcer{},
// 				ReferrerStores: []referrerstore.ReferrerStore{},
// 				Verifiers:      []verifier.ReferenceVerifier{},
// 				Config: &exConfig.ExecutorConfig{
// 					VerificationRequestTimeout: new(int),
// 					MutationRequestTimeout:     nil,
// 				},
// 			},
// 			expectedTimeout: 5000,
// 		},
// 	}

// 	for _, testcase := range testcases {
// 		if testcase.setTimeout >= 0 {
// 			*testcase.ex.Config.VerificationRequestTimeout = testcase.setTimeout
// 		}
// 		expected := time.Millisecond * time.Duration(testcase.expectedTimeout)
// 		actual := testcase.ex.GetVerifyRequestTimeout()
// 		if actual != expected {
// 			t.Fatalf("verification request timeout returned expected %dms but got %dms", expected.Milliseconds(), actual.Milliseconds())
// 		}
// 	}
// }

// // TestGetMutationRequestTimeout_ExpectedResults tests the mutation request timeout returned
// func TestGetMutationRequestTimeout_ExpectedResults(t *testing.T) {
// 	testcases := []struct {
// 		setTimeout      int
// 		ex              Executor
// 		expectedTimeout int
// 	}{
// 		{
// 			setTimeout: -1,
// 			ex: Executor{
// 				PolicyEnforcer: policyConfig.PolicyEnforcer{},
// 				ReferrerStores: []referrerstore.ReferrerStore{},
// 				Verifiers:      []verifier.ReferenceVerifier{},
// 				Config:         nil,
// 			},
// 			expectedTimeout: 950,
// 		},
// 		{
// 			setTimeout: -1,
// 			ex: Executor{
// 				PolicyEnforcer: policyConfig.PolicyEnforcer{},
// 				ReferrerStores: []referrerstore.ReferrerStore{},
// 				Verifiers:      []verifier.ReferenceVerifier{},
// 				Config: &exConfig.ExecutorConfig{
// 					VerificationRequestTimeout: nil,
// 					MutationRequestTimeout:     nil,
// 				},
// 			},
// 			expectedTimeout: 950,
// 		},
// 		{
// 			setTimeout: 2400,
// 			ex: Executor{
// 				PolicyEnforcer: policyConfig.PolicyEnforcer{},
// 				ReferrerStores: []referrerstore.ReferrerStore{},
// 				Verifiers:      []verifier.ReferenceVerifier{},
// 				Config: &exConfig.ExecutorConfig{
// 					VerificationRequestTimeout: nil,
// 					MutationRequestTimeout:     new(int),
// 				},
// 			},
// 			expectedTimeout: 2400,
// 		},
// 	}

// 	for _, testcase := range testcases {
// 		if testcase.setTimeout >= 0 {
// 			*testcase.ex.Config.MutationRequestTimeout = testcase.setTimeout
// 		}
// 		expected := time.Millisecond * time.Duration(testcase.expectedTimeout)
// 		actual := testcase.ex.GetMutationRequestTimeout()
// 		if actual != expected {
// 			t.Fatalf("mutation request timeout returned expected %dms but got %dms", expected.Milliseconds(), actual.Milliseconds())
// 		}
// 	}
// }

// func TestVerifySubject(t *testing.T) {
// 	testCases := []struct {
// 		name           string
// 		params         e.VerifyParameters
// 		expectedResult types.VerifyResult
// 		stores         []referrerstore.ReferrerStore
// 		policyEnforcer policyprovider.PolicyProvider
// 		verifiers      []verifier.ReferenceVerifier
// 		config         *exConfig.ExecutorConfig
// 		referrers      []ocispecs.ReferenceDescriptor
// 		expectErr      bool
// 	}{
// 		{
// 			name: "verify subject with invalid subject",
// 			config: &exConfig.ExecutorConfig{
// 				UseRegoPolicy: true,
// 			},
// 			expectErr: true,
// 		},
// 		{
// 			name: "error from ListReferrers",
// 			config: &exConfig.ExecutorConfig{
// 				UseRegoPolicy: true,
// 			},
// 			params: e.VerifyParameters{
// 				Subject: subject1,
// 			},
// 			stores: []referrerstore.ReferrerStore{
// 				&mockStore{
// 					referrers: nil,
// 				},
// 			},
// 			expectErr:      true,
// 			expectedResult: types.VerifyResult{},
// 		},
// 		{
// 			name: "empty referrers",
// 			config: &exConfig.ExecutorConfig{
// 				UseRegoPolicy: true,
// 			},
// 			params: e.VerifyParameters{
// 				Subject: subject1,
// 			},
// 			stores: []referrerstore.ReferrerStore{
// 				&mockStore{
// 					referrers: nil,
// 				},
// 			},
// 			expectErr:      true,
// 			expectedResult: types.VerifyResult{},
// 		},
// 		{
// 			name: "one signature without matching verifier",
// 			config: &exConfig.ExecutorConfig{
// 				UseRegoPolicy: true,
// 			},
// 			params: e.VerifyParameters{
// 				Subject: subject1,
// 			},
// 			stores: []referrerstore.ReferrerStore{
// 				&mockStore{
// 					referrers: map[string][]ocispecs.ReferenceDescriptor{
// 						subjectDigest: {
// 							{
// 								ArtifactType: artifactType,
// 								Descriptor: oci.Descriptor{
// 									Digest: signatureDigest,
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			verifiers: []verifier.ReferenceVerifier{
// 				&mockVerifier{
// 					canVerify: false,
// 				},
// 			},
// 			policyEnforcer: &mockPolicyProvider{result: true},
// 			expectErr:      false,
// 			expectedResult: types.VerifyResult{
// 				IsSuccess: true,
// 			},
// 		},
// 		{
// 			name: "one signature with verifier failed",
// 			config: &exConfig.ExecutorConfig{
// 				UseRegoPolicy: true,
// 			},
// 			params: e.VerifyParameters{
// 				Subject: subject1,
// 			},
// 			stores: []referrerstore.ReferrerStore{
// 				&mockStore{
// 					referrers: map[string][]ocispecs.ReferenceDescriptor{
// 						subjectDigest: {
// 							{
// 								ArtifactType: artifactType,
// 								Descriptor: oci.Descriptor{
// 									Digest: signatureDigest,
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			verifiers: []verifier.ReferenceVerifier{
// 				&mockVerifier{
// 					canVerify: true,
// 					verifierResult: verifier.VerifierResult{
// 						IsSuccess: false,
// 					},
// 				},
// 			},
// 			policyEnforcer: &mockPolicyProvider{result: false},
// 			expectErr:      false,
// 			expectedResult: types.VerifyResult{IsSuccess: false},
// 		},
// 		{
// 			name: "one signature with verifier success",
// 			config: &exConfig.ExecutorConfig{
// 				UseRegoPolicy: true,
// 			},
// 			params: e.VerifyParameters{
// 				Subject: subject1,
// 			},
// 			stores: []referrerstore.ReferrerStore{
// 				&mockStore{
// 					referrers: map[string][]ocispecs.ReferenceDescriptor{
// 						subjectDigest: {
// 							{
// 								ArtifactType: artifactType,
// 								Descriptor: oci.Descriptor{
// 									Digest: signatureDigest,
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			verifiers: []verifier.ReferenceVerifier{
// 				&mockVerifier{
// 					canVerify: true,
// 					verifierResult: verifier.VerifierResult{
// 						IsSuccess: true,
// 					},
// 				},
// 			},
// 			policyEnforcer: &mockPolicyProvider{result: true},
// 			expectErr:      false,
// 			expectedResult: types.VerifyResult{
// 				IsSuccess: true,
// 			},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			ex, err := NewExecutor(tc.stores, tc.policyEnforcer, tc.verifiers, tc.config)
// 			if err != nil {
// 				t.Fatalf("failed to create executor: %v", err)
// 			}

// 			result, err := ex.VerifySubject(context.Background(), tc.params)
// 			if (err != nil) != tc.expectErr {
// 				t.Fatalf("expected error %v but got %v", tc.expectErr, err)
// 			}
// 			if result.IsSuccess != tc.expectedResult.IsSuccess {
// 				t.Fatalf("expected result: %+v but got: %+v", tc.expectedResult, result)
// 			}
// 		})
// 	}
// }
