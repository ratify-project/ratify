package policyprovider

import (
	"context"

	"github.com/deislabs/hora/pkg/common"
	"github.com/deislabs/hora/pkg/executor/types"
	"github.com/deislabs/hora/pkg/ocispecs"
)

type PolicyProvider interface {
	VerifyNeeded(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor) bool
	ContinueVerifyOnFailure(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor, partialVerifyResult types.VerifyResult) bool
	// which errors to treat as failure ?
	ErrorToVerifyResult(ctx context.Context, subjectRefString string, verifyError error) types.VerifyResult
}
