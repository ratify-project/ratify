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
