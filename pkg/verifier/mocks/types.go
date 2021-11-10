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

package mocks

import (
	"context"

	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/executor"
	"github.com/deislabs/ratify/pkg/executor/types"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore"
	"github.com/deislabs/ratify/pkg/referrerstore/config"
	"github.com/deislabs/ratify/pkg/verifier"
	"github.com/opencontainers/go-digest"
)

type TestStore struct{}

func (s *TestStore) Name() string {
	return "test-store"
}

func (s *TestStore) ListReferrers(ctx context.Context, subjectReference common.Reference, artifactTypes []string, nextToken string) (referrerstore.ListReferrersResult, error) {
	return referrerstore.ListReferrersResult{}, nil
}

func (s *TestStore) GetBlobContent(ctx context.Context, subjectReference common.Reference, digest digest.Digest) ([]byte, error) {
	return nil, nil
}

func (s *TestStore) GetReferenceManifest(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor) (ocispecs.ReferenceManifest, error) {
	return ocispecs.ReferenceManifest{}, nil
}

func (s *TestStore) GetConfig() *config.StoreConfig {
	return &config.StoreConfig{}
}

type TestExecutor struct {
	VerifySuccess bool
}

func (s *TestExecutor) VerifySubject(ctx context.Context, verifyParameters executor.VerifyParameters) (types.VerifyResult, error) {
	report := verifier.VerifierResult{IsSuccess: s.VerifySuccess}
	return types.VerifyResult{
		IsSuccess:       s.VerifySuccess,
		VerifierReports: []interface{}{report}}, nil
}
