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

	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore"
	"github.com/deislabs/ratify/pkg/verifier"
)

type TestVerifier struct {
	CanVerifyFunc func(artifactType string) bool
	VerifyResult  func(artifactType string) bool
}

func (s *TestVerifier) Name() string {
	return "test-verifier"
}

func (s *TestVerifier) CanVerify(ctx context.Context, referenceDescriptor ocispecs.ReferenceDescriptor) bool {
	return s.CanVerifyFunc(referenceDescriptor.ArtifactType)
}

func (s *TestVerifier) Verify(ctx context.Context,
	subjectReference common.Reference,
	referenceDescriptor ocispecs.ReferenceDescriptor,
	referrerStore referrerstore.ReferrerStore) (verifier.VerifierResult, error) {
	return verifier.VerifierResult{
		IsSuccess: s.VerifyResult(referenceDescriptor.ArtifactType),
	}, nil
}
