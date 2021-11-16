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
	"testing"

	e "github.com/deislabs/ratify/pkg/executor"
	"github.com/deislabs/ratify/pkg/ocispecs"
	config "github.com/deislabs/ratify/pkg/policyprovider/configpolicy"
	"github.com/deislabs/ratify/pkg/referrerstore"
	"github.com/deislabs/ratify/pkg/verifier"
	"github.com/opencontainers/go-digest"
)

func TestVerifySubject_SubjectParseError(t *testing.T) {
	executor := Executor{}

	verifyParameters := e.VerifyParameters{
		Subject: "localhost:5000/net-monitor:v1",
	}

	_, err := executor.verifySubjectInternal(context.Background(), verifyParameters)

	if err == nil {
		t.Fatal("expected subject parsing to fail")
	}
}

func TestVerifySubject_ResolveTag_Success(t *testing.T) {
	testDigest := digest.FromString("test")
	store := &TestStore{
		references: []ocispecs.ReferenceDescriptor{},
		resolveMap: map[string]digest.Digest{
			"v1": testDigest,
		},
	}

	executor := Executor{
		ReferrerStores: []referrerstore.ReferrerStore{store},
	}

	verifyParameters := e.VerifyParameters{
		Subject: "localhost:5000/net-monitor:v1",
	}

	_, err := executor.verifySubjectInternal(context.Background(), verifyParameters)

	if err != ReferrersNotFound {
		t.Fatalf("expected ReferrersNotFound actual %v", err)
	}
}

func TestVerifySubject_Verify_NoReferrers(t *testing.T) {
	configPolicy := config.PolicyEnforcer{}
	ex := &Executor{
		PolicyEnforcer: configPolicy,
		ReferrerStores: []referrerstore.ReferrerStore{&TestStore{}},
		Verifiers:      []verifier.ReferenceVerifier{&TestVerifier{}},
	}

	verifyParameters := e.VerifyParameters{
		Subject: "localhost:5000/net-monitor:v1@sha256:a0fc570a245b09ed752c42d600ee3bb5b4f77bbd70d8898780b7ab43454530eb",
	}

	_, err := ex.verifySubjectInternal(context.Background(), verifyParameters)

	if err != ReferrersNotFound {
		t.Fatalf("expected ReferrersNotFound actual %v", err)
	}
}

func TestVerifySubject_CanVerify_ExpectedResults(t *testing.T) {
	configPolicy := config.PolicyEnforcer{}
	store := &TestStore{references: []ocispecs.ReferenceDescriptor{
		{
			ArtifactType: "test-type1",
		},
		{
			ArtifactType: "test-type2",
		},
	}}
	ver := &TestVerifier{
		canVerify: func(at string) bool {
			return at == "test-type1"
		},
		verifyResult: func(artifactType string) bool {
			return true
		},
	}

	ex := &Executor{
		PolicyEnforcer: configPolicy,
		ReferrerStores: []referrerstore.ReferrerStore{store},
		Verifiers:      []verifier.ReferenceVerifier{ver},
	}

	verifyParameters := e.VerifyParameters{
		Subject: "localhost:5000/net-monitor:v1@sha256:a0fc570a245b09ed752c42d600ee3bb5b4f77bbd70d8898780b7ab43454530eb",
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
	configPolicy := config.PolicyEnforcer{}
	store := &TestStore{references: []ocispecs.ReferenceDescriptor{
		{
			ArtifactType: "test-type1",
		},
		{
			ArtifactType: "test-type2",
		},
	}}
	ver := &TestVerifier{
		canVerify: func(at string) bool {
			return true
		},
		verifyResult: func(artifactType string) bool {
			if artifactType == "test-type1" {
				return false
			}

			return true
		},
	}

	ex := &Executor{
		PolicyEnforcer: configPolicy,
		ReferrerStores: []referrerstore.ReferrerStore{store},
		Verifiers:      []verifier.ReferenceVerifier{ver},
	}

	verifyParameters := e.VerifyParameters{
		Subject: "localhost:5000/net-monitor:v1@sha256:a0fc570a245b09ed752c42d600ee3bb5b4f77bbd70d8898780b7ab43454530eb",
	}

	result, err := ex.verifySubjectInternal(context.Background(), verifyParameters)

	if err != nil {
		t.Fatalf("verification failed with err %v", err)
	}

	if result.IsSuccess {
		t.Fatal("verification expected to fail")
	}

	if len(result.VerifierReports) != 1 {
		t.Fatalf("verification expected to return single report but actual count %d", len(result.VerifierReports))
	}
}

func TestVerifySubject_VerifySuccess_ExpectedResults(t *testing.T) {
	configPolicy := config.PolicyEnforcer{}
	store := &TestStore{references: []ocispecs.ReferenceDescriptor{
		{
			ArtifactType: "test-type1",
		},
		{
			ArtifactType: "test-type2",
		},
	}}
	ver := &TestVerifier{
		canVerify: func(at string) bool {
			return true
		},
		verifyResult: func(artifactType string) bool {
			return true
		},
	}

	ex := &Executor{
		PolicyEnforcer: configPolicy,
		ReferrerStores: []referrerstore.ReferrerStore{store},
		Verifiers:      []verifier.ReferenceVerifier{ver},
	}

	verifyParameters := e.VerifyParameters{
		Subject: "localhost:5000/net-monitor:v1@sha256:a0fc570a245b09ed752c42d600ee3bb5b4f77bbd70d8898780b7ab43454530eb",
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
