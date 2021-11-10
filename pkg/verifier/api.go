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

package verifier

import (
	"context"

	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/executor"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore"
)

type VerifierResult struct {
	Subject       string           `json:"subject,omitempty"`
	IsSuccess     bool             `json:"isSuccess,omitempty"`
	Name          string           `json:"name,omitempty"`
	Results       []string         `json:"results,omitempty"`
	NestedResults []VerifierResult `json:"nestedResults,omitempty"`
}

type ReferenceVerifier interface {
	Name() string
	CanVerify(ctx context.Context, referenceDescriptor ocispecs.ReferenceDescriptor) bool
	Verify(ctx context.Context,
		subjectReference common.Reference,
		referenceDescriptor ocispecs.ReferenceDescriptor,
		referrerStore referrerstore.ReferrerStore,
		executor executor.Executor) (VerifierResult, error)
}
