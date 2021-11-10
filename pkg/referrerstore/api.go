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

package referrerstore

import (
	"context"

	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore/config"
	"github.com/opencontainers/go-digest"
)

type ListReferrersResult struct {
	Referrers []ocispecs.ReferenceDescriptor
	NextToken string
}

type ReferrerStore interface {
	Name() string
	ListReferrers(ctx context.Context, subjectReference common.Reference, artifactTypes []string, nextToken string) (ListReferrersResult, error)
	// Used for small objects.
	GetBlobContent(ctx context.Context, subjectReference common.Reference, digest digest.Digest) ([]byte, error)
	GetReferenceManifest(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor) (ocispecs.ReferenceManifest, error)
	GetConfig() *config.StoreConfig
}
