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
