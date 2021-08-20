package policyprovider

import (
	"context"

	"github.com/notaryproject/hora/pkg/common"
	"github.com/notaryproject/hora/pkg/executor/types"
	"github.com/notaryproject/hora/pkg/ocispecs"
)

type PolicyProvider interface {
	VerifyNeeded(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor) bool
	ContinueVerifyOnFailure(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor, partialVerifyResult types.VerifyResult) bool
	// which errors to treat as failure ?
	ErrorToVerifyResult(ctx context.Context, subjectRefString string, verifyError error) types.VerifyResult
}
