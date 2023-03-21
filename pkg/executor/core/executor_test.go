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

import (
	"context"
	"errors"
	"testing"
	"time"

	exConfig "github.com/deislabs/ratify/pkg/executor/config"

	e "github.com/deislabs/ratify/pkg/executor"
	"github.com/deislabs/ratify/pkg/ocispecs"
	config "github.com/deislabs/ratify/pkg/policyprovider/configpolicy"
	"github.com/deislabs/ratify/pkg/policyprovider/types"
	"github.com/deislabs/ratify/pkg/referrerstore"
	"github.com/deislabs/ratify/pkg/referrerstore/mocks"
	"github.com/deislabs/ratify/pkg/verifier"
	"github.com/opencontainers/go-digest"
)

const (
	testArtifactType1 = "test-type1"
	testArtifactType2 = "test-type2"
)

func TestVerifySubject_ResolveSubjectDescriptor_Failed(t *testing.T) {
	executor := Executor{}

	verifyParameters := e.VerifyParameters{
		Subject: "localhost:5000/net-monitor:v1",
	}

	_, err := executor.verifySubjectInternal(context.Background(), verifyParameters)

	if err == nil {
		t.Fatal("expected subject parsing to fail")
	}
}

func TestVerifySubject_ResolveSubjectDescriptor_Success(t *testing.T) {
	testDigest := digest.FromString("test")
	store := &mocks.TestStore{
		References: []ocispecs.ReferenceDescriptor{},
		ResolveMap: map[string]digest.Digest{
			"v1": testDigest,
		},
	}

	executor := Executor{
		ReferrerStores: []referrerstore.ReferrerStore{store},
	}

	verifyParameters := e.VerifyParameters{
		Subject: "localhost:5000/net-monitor:v1",
	}

	if _, err := executor.verifySubjectInternal(context.Background(), verifyParameters); !errors.Is(err, ErrReferrersNotFound) {
		t.Fatalf("expected ErrReferrersNotFound actual %v", err)
	}
}

func TestVerifySubject_Verify_NoReferrers(t *testing.T) {
	testDigest := digest.FromString("test")
	configPolicy := config.PolicyEnforcer{}
	ex := &Executor{
		PolicyEnforcer: configPolicy,
		ReferrerStores: []referrerstore.ReferrerStore{&mocks.TestStore{
			ResolveMap: map[string]digest.Digest{
				"v1": testDigest,
			},
		}},
		Verifiers: []verifier.ReferenceVerifier{&TestVerifier{}},
		Config: &exConfig.ExecutorConfig{
			VerificationRequestTimeout: nil,
			MutationRequestTimeout:     nil,
		},
	}

	verifyParameters := e.VerifyParameters{
		Subject: "localhost:5000/net-monitor:v1",
	}

	if _, err := ex.verifySubjectInternal(context.Background(), verifyParameters); !errors.Is(err, ErrReferrersNotFound) {
		t.Fatalf("expected ErrReferrersNotFound actual %v", err)
	}
}

func TestVerifySubject_CanVerify_ExpectedResults(t *testing.T) {
	testDigest := digest.FromString("test")
	configPolicy := config.PolicyEnforcer{
		ArtifactTypePolicies: map[string]types.ArtifactTypeVerifyPolicy{
			testArtifactType1: types.AnyVerifySuccess,
		}}
	store := &mocks.TestStore{References: []ocispecs.ReferenceDescriptor{
		{
			ArtifactType: testArtifactType1,
		},
		{
			ArtifactType: "test-type2",
		}},
		ResolveMap: map[string]digest.Digest{
			"v1": testDigest,
		},
	}
	ver := &TestVerifier{
		CanVerifyFunc: func(at string) bool {
			return at == testArtifactType1
		},
		VerifyResult: func(artifactType string) bool {
			return true
		},
	}

	ex := &Executor{
		PolicyEnforcer: configPolicy,
		ReferrerStores: []referrerstore.ReferrerStore{store},
		Verifiers:      []verifier.ReferenceVerifier{ver},
		Config: &exConfig.ExecutorConfig{
			VerificationRequestTimeout: nil,
			MutationRequestTimeout:     nil,
		},
	}

	verifyParameters := e.VerifyParameters{
		Subject: "localhost:5000/net-monitor:v1",
	}

	result, err := ex.verifySubjectInternal(context.Background(), verifyParameters)

	if err != nil {
		t.Fatalf("verification failed with err %v", err)
	}

	if !result.IsSuccess {
		t.Fatal("verification expected to be success")
	}

	if len(result.VerifierReports) != 1 {
		t.Fatalf("verification expected to return single report but actual count %d", len(result.VerifierReports))
	}
}

func TestVerifySubject_VerifyFailures_ExpectedResults(t *testing.T) {
	testDigest := digest.FromString("test")
	testArtifactType := "test-type1"
	configPolicy := config.PolicyEnforcer{
		ArtifactTypePolicies: map[string]types.ArtifactTypeVerifyPolicy{
			testArtifactType: types.AnyVerifySuccess,
		}}
	store := &mocks.TestStore{References: []ocispecs.ReferenceDescriptor{
		{
			ArtifactType: testArtifactType,
		},
		{
			ArtifactType: "test-type2",
		}},
		ResolveMap: map[string]digest.Digest{
			"v1": testDigest,
		},
	}
	ver := &TestVerifier{
		CanVerifyFunc: func(at string) bool {
			return true
		},
		VerifyResult: func(artifactType string) bool {
			return artifactType != testArtifactType
		},
	}

	ex := &Executor{
		PolicyEnforcer: configPolicy,
		ReferrerStores: []referrerstore.ReferrerStore{store},
		Verifiers:      []verifier.ReferenceVerifier{ver},
		Config: &exConfig.ExecutorConfig{
			VerificationRequestTimeout: nil,
			MutationRequestTimeout:     nil,
		},
	}

	verifyParameters := e.VerifyParameters{
		Subject: "localhost:5000/net-monitor:v1",
	}

	result, err := ex.verifySubjectInternal(context.Background(), verifyParameters)

	if err != nil {
		t.Fatalf("verification failed with err %v", err)
	}

	if result.IsSuccess {
		t.Fatal("verification expected to fail")
	}
}

func TestVerifySubject_VerifySuccess_ExpectedResults(t *testing.T) {
	testDigest := digest.FromString("test")
	configPolicy := config.PolicyEnforcer{
		ArtifactTypePolicies: map[string]types.ArtifactTypeVerifyPolicy{
			testArtifactType1: types.AnyVerifySuccess,
			testArtifactType2: types.AnyVerifySuccess,
		}}
	store := &mocks.TestStore{References: []ocispecs.ReferenceDescriptor{
		{
			ArtifactType: testArtifactType1,
		},
		{
			ArtifactType: testArtifactType2,
		}},
		ResolveMap: map[string]digest.Digest{
			"v1": testDigest,
		},
	}
	ver := &TestVerifier{
		CanVerifyFunc: func(at string) bool {
			return true
		},
		VerifyResult: func(artifactType string) bool {
			return true
		},
	}

	ex := &Executor{
		PolicyEnforcer: configPolicy,
		ReferrerStores: []referrerstore.ReferrerStore{store},
		Verifiers:      []verifier.ReferenceVerifier{ver},
		Config: &exConfig.ExecutorConfig{
			VerificationRequestTimeout: nil,
			MutationRequestTimeout:     nil,
		},
	}

	verifyParameters := e.VerifyParameters{
		Subject: "localhost:5000/net-monitor:v1",
	}

	result, err := ex.verifySubjectInternal(context.Background(), verifyParameters)

	if err != nil {
		t.Fatalf("verification failed with err %v", err)
	}

	if !result.IsSuccess {
		t.Fatal("verification expected to fail")
	}

	if len(result.VerifierReports) != 2 {
		t.Fatalf("verification expected to return two reports but actual count %d", len(result.VerifierReports))
	}
}

// TestVerifySubject_MultipleArtifacts_ExpectedResults tests multiple artifacts are verified concurrently
func TestVerifySubject_MultipleArtifacts_ExpectedResults(t *testing.T) {
	testDigest := digest.FromString("test")
	configPolicy := config.PolicyEnforcer{
		ArtifactTypePolicies: map[string]types.ArtifactTypeVerifyPolicy{
			testArtifactType1: types.AnyVerifySuccess,
			testArtifactType2: types.AnyVerifySuccess,
		}}
	store := &mocks.TestStore{References: []ocispecs.ReferenceDescriptor{
		{
			ArtifactType: testArtifactType1,
		},
		{
			ArtifactType: testArtifactType2,
		}},
		ResolveMap: map[string]digest.Digest{
			"v1": testDigest,
		},
	}
	ver := &TestVerifier{
		CanVerifyFunc: func(at string) bool {
			return true
		},
		VerifyResult: func(artifactType string) bool {
			if artifactType == testArtifactType1 {
				time.Sleep(2 * time.Second)
			}
			return true
		},
	}

	ex := &Executor{
		PolicyEnforcer: configPolicy,
		ReferrerStores: []referrerstore.ReferrerStore{store},
		Verifiers:      []verifier.ReferenceVerifier{ver},
		Config: &exConfig.ExecutorConfig{
			VerificationRequestTimeout: nil,
			MutationRequestTimeout:     nil,
		},
	}

	verifyParameters := e.VerifyParameters{
		Subject: "localhost:5000/net-monitor:v1",
	}

	result, err := ex.verifySubjectInternal(context.Background(), verifyParameters)

	if err != nil {
		t.Fatalf("verification failed with err %v", err)
	}

	if !result.IsSuccess {
		t.Fatal("verification expected to fail")
	}

	if len(result.VerifierReports) != 2 {
		t.Fatalf("verification expected to return two reports but actual count %d", len(result.VerifierReports))
	}

	if result.VerifierReports[0].(verifier.VerifierResult).ArtifactType != "test-type2" {
		t.Fatalf("verification expected to return second artifact verifier report first")
	}
}

// TestVerifySubject_NestedReferences_Expected tests verifier config can specify nested references
func TestVerifySubject_NestedReferences_Expected(t *testing.T) {
	configPolicy := config.PolicyEnforcer{
		ArtifactTypePolicies: map[string]types.ArtifactTypeVerifyPolicy{
			"default": "all",
		}}

	store := mocks.CreateNewTestStoreForNestedSbom()

	// sbom verifier WITH nested references in config
	sbomVerifier := &TestVerifier{
		CanVerifyFunc: func(at string) bool {
			return at == mocks.SbomArtifactType
		},
		VerifyResult: func(artifactType string) bool {
			return true
		},
		nestedReferences: []string{"string-content-does-not-matter"},
	}

	signatureVerifier := &TestVerifier{
		CanVerifyFunc: func(at string) bool {
			return at == mocks.SignatureArtifactType
		},
		VerifyResult: func(artifactType string) bool {
			return true
		},
	}

	ex := &Executor{
		PolicyEnforcer: configPolicy,
		ReferrerStores: []referrerstore.ReferrerStore{store},
		Verifiers:      []verifier.ReferenceVerifier{sbomVerifier, signatureVerifier},
		Config: &exConfig.ExecutorConfig{
			VerificationRequestTimeout: nil,
			MutationRequestTimeout:     nil,
		},
	}

	verifyParameters := e.VerifyParameters{
		Subject: mocks.TestSubjectWithDigest,
	}

	result, err := ex.verifySubjectInternal(context.Background(), verifyParameters)

	if err != nil {
		t.Fatalf("verification failed with err %v", err)
	}

	if !result.IsSuccess {
		t.Fatal("verification expected to succeed")
	}

	if len(result.VerifierReports) != 2 {
		t.Fatalf("verification expected to return two reports but actual count %d", len(result.VerifierReports))
	}

	for _, report := range result.VerifierReports {
		castedReport := report.(verifier.VerifierResult)

		// check sbom report
		if castedReport.ArtifactType == mocks.SbomArtifactType {
			// check sbom has one nested results
			if len(castedReport.NestedResults) != 1 {
				t.Fatalf("Expected sbom report to have 1 nested result")
			}
			// check sbom nested result is successful
			if !castedReport.NestedResults[0].IsSuccess {
				t.Fatalf("Expected the sbom nested result to be successful")
			}
		} else {
			// check non-sbom reports have zero nested results
			if len(castedReport.NestedResults) != 0 {
				t.Fatalf("Expected non-sboms reports to have zero nested results")
			}
		}
	}
}

// TestVerifySubject__NoNestedReferences_Expected tests verifier config can specify no nested references
func TestVerifySubject_NoNestedReferences_Expected(t *testing.T) {
	configPolicy := config.PolicyEnforcer{
		ArtifactTypePolicies: map[string]types.ArtifactTypeVerifyPolicy{
			"default": "all",
		}}
	store := mocks.CreateNewTestStoreForNestedSbom()

	// sbom verifier WITHOUT nested references in config
	sbomVer := &TestVerifier{
		CanVerifyFunc: func(at string) bool {
			return at == mocks.SbomArtifactType
		},
		VerifyResult: func(artifactType string) bool {
			return true
		},
	}

	signatureVer := &TestVerifier{
		CanVerifyFunc: func(at string) bool {
			return at == mocks.SignatureArtifactType
		},
		VerifyResult: func(artifactType string) bool {
			return true
		},
	}

	ex := &Executor{
		PolicyEnforcer: configPolicy,
		ReferrerStores: []referrerstore.ReferrerStore{store},
		Verifiers:      []verifier.ReferenceVerifier{sbomVer, signatureVer},
		Config: &exConfig.ExecutorConfig{
			VerificationRequestTimeout: nil,
			MutationRequestTimeout:     nil,
		},
	}

	verifyParameters := e.VerifyParameters{
		Subject: mocks.TestSubjectWithDigest,
	}

	result, err := ex.verifySubjectInternal(context.Background(), verifyParameters)

	if err != nil {
		t.Fatalf("verification failed with err %v", err)
	}

	if !result.IsSuccess {
		t.Fatal("verification expected to succeed")
	}

	if len(result.VerifierReports) != 2 {
		t.Fatalf("verification expected to return two reports but actual count %d", len(result.VerifierReports))
	}

	// check each report for: success, zero nested results
	for _, report := range result.VerifierReports {
		castedReport := report.(verifier.VerifierResult)

		// check for success
		if !castedReport.IsSuccess {
			t.Fatal("verification expected to succeed")
		}
		// check there are no nested results
		if len(castedReport.NestedResults) != 0 {
			t.Fatalf("expected reports to have zero nested results")
		}
	}
}

// TestGetVerifyRequestTimeout_ExpectedResults tests the verification request timeout returned
func TestGetVerifyRequestTimeout_ExpectedResults(t *testing.T) {
	testcases := []struct {
		setTimeout      int
		ex              Executor
		expectedTimeout int
	}{
		{
			setTimeout: -1,
			ex: Executor{
				PolicyEnforcer: config.PolicyEnforcer{},
				ReferrerStores: []referrerstore.ReferrerStore{},
				Verifiers:      []verifier.ReferenceVerifier{},
				Config:         nil,
			},
			expectedTimeout: 2900,
		},
		{
			setTimeout: -1,
			ex: Executor{
				PolicyEnforcer: config.PolicyEnforcer{},
				ReferrerStores: []referrerstore.ReferrerStore{},
				Verifiers:      []verifier.ReferenceVerifier{},
				Config: &exConfig.ExecutorConfig{
					VerificationRequestTimeout: nil,
					MutationRequestTimeout:     nil,
				},
			},
			expectedTimeout: 2900,
		},
		{
			setTimeout: 5000,
			ex: Executor{
				PolicyEnforcer: config.PolicyEnforcer{},
				ReferrerStores: []referrerstore.ReferrerStore{},
				Verifiers:      []verifier.ReferenceVerifier{},
				Config: &exConfig.ExecutorConfig{
					VerificationRequestTimeout: new(int),
					MutationRequestTimeout:     nil,
				},
			},
			expectedTimeout: 5000,
		},
	}

	for _, testcase := range testcases {
		if testcase.setTimeout >= 0 {
			*testcase.ex.Config.VerificationRequestTimeout = testcase.setTimeout
		}
		expected := time.Millisecond * time.Duration(testcase.expectedTimeout)
		actual := testcase.ex.GetVerifyRequestTimeout()
		if actual != expected {
			t.Fatalf("verification request timeout returned expected %dms but got %dms", expected.Milliseconds(), actual.Milliseconds())
		}
	}
}

// TestGetMutationRequestTimeout_ExpectedResults tests the mutation request timeout returned
func TestGetMutationRequestTimeout_ExpectedResults(t *testing.T) {
	testcases := []struct {
		setTimeout      int
		ex              Executor
		expectedTimeout int
	}{
		{
			setTimeout: -1,
			ex: Executor{
				PolicyEnforcer: config.PolicyEnforcer{},
				ReferrerStores: []referrerstore.ReferrerStore{},
				Verifiers:      []verifier.ReferenceVerifier{},
				Config:         nil,
			},
			expectedTimeout: 950,
		},
		{
			setTimeout: -1,
			ex: Executor{
				PolicyEnforcer: config.PolicyEnforcer{},
				ReferrerStores: []referrerstore.ReferrerStore{},
				Verifiers:      []verifier.ReferenceVerifier{},
				Config: &exConfig.ExecutorConfig{
					VerificationRequestTimeout: nil,
					MutationRequestTimeout:     nil,
				},
			},
			expectedTimeout: 950,
		},
		{
			setTimeout: 2400,
			ex: Executor{
				PolicyEnforcer: config.PolicyEnforcer{},
				ReferrerStores: []referrerstore.ReferrerStore{},
				Verifiers:      []verifier.ReferenceVerifier{},
				Config: &exConfig.ExecutorConfig{
					VerificationRequestTimeout: nil,
					MutationRequestTimeout:     new(int),
				},
			},
			expectedTimeout: 2400,
		},
	}

	for _, testcase := range testcases {
		if testcase.setTimeout >= 0 {
			*testcase.ex.Config.MutationRequestTimeout = testcase.setTimeout
		}
		expected := time.Millisecond * time.Duration(testcase.expectedTimeout)
		actual := testcase.ex.GetMutationRequestTimeout()
		if actual != expected {
			t.Fatalf("mutation request timeout returned expected %dms but got %dms", expected.Milliseconds(), actual.Milliseconds())
		}
	}
}
