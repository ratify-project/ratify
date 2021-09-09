package verifier

import (
	"context"

	"github.com/deislabs/hora/pkg/common"
	"github.com/deislabs/hora/pkg/executor"
	"github.com/deislabs/hora/pkg/ocispecs"
	rc "github.com/deislabs/hora/pkg/referrerstore/config"
)

type VerifierResult struct {
	Subject   string
	IsSuccess bool
	Name      string
	Results   []string
}

type ReferenceVerifier interface {
	Name() string
	CanVerify(ctx context.Context, referenceDescriptor ocispecs.ReferenceDescriptor) bool
	Verify(ctx context.Context,
		subjectReference common.Reference,
		referenceDescriptor ocispecs.ReferenceDescriptor,
		referrerStoreConfig *rc.StoreConfig,
		executor executor.Executor) (VerifierResult, error)
}
