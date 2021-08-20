package verifier

import (
	"context"

	"github.com/notaryproject/hora/pkg/common"
	"github.com/notaryproject/hora/pkg/executor"
	"github.com/notaryproject/hora/pkg/ocispecs"
	"github.com/notaryproject/hora/pkg/referrerstore"
)

type VerifierResult struct {
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
		referrerStore referrerstore.ReferrerStore,
		executor executor.Executor) (VerifierResult, error)
}
