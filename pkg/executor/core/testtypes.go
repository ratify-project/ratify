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
	"fmt"

	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/executor"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore"
	"github.com/deislabs/ratify/pkg/referrerstore/config"
	"github.com/deislabs/ratify/pkg/verifier"
	"github.com/opencontainers/go-digest"
)

type TestStore struct {
	references []ocispecs.ReferenceDescriptor
	resolveMap map[string]digest.Digest
}

func (s *TestStore) Name() string {
	return "test-store"
}

func (s *TestStore) ListReferrers(ctx context.Context, subjectReference common.Reference, artifactTypes []string, nextToken string) (referrerstore.ListReferrersResult, error) {
	return referrerstore.ListReferrersResult{Referrers: s.references}, nil
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

func (s *TestStore) ResolveTag(ctx context.Context, subjectReference common.Reference) (digest.Digest, error) {
	if s.resolveMap != nil {
		if result, ok := s.resolveMap[subjectReference.Tag]; ok {
			return result, nil
		}
	}

	return "", fmt.Errorf("cannot resolve digest for the subject reference")
}

type TestVerifier struct {
	canVerify    func(artifactType string) bool
	verifyResult func(artifactType string) bool
}

func (s *TestVerifier) Name() string {
	return "test-verifier"
}

func (s *TestVerifier) CanVerify(ctx context.Context, referenceDescriptor ocispecs.ReferenceDescriptor) bool {
	return s.canVerify(referenceDescriptor.ArtifactType)
}

func (s *TestVerifier) Verify(ctx context.Context,
	subjectReference common.Reference,
	referenceDescriptor ocispecs.ReferenceDescriptor,
	referrerStore referrerstore.ReferrerStore,
	executor executor.Executor) (verifier.VerifierResult, error) {
	return verifier.VerifierResult{
		IsSuccess: s.verifyResult(referenceDescriptor.ArtifactType),
	}, nil
}
