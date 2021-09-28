package config

import (
	"context"
	"fmt"

	"github.com/deislabs/hora/pkg/common"
	"github.com/deislabs/hora/pkg/executor/types"
	"github.com/deislabs/hora/pkg/ocispecs"
	"github.com/deislabs/hora/pkg/verifier"
)

type PolicyEnforcer struct {
}

func (enforcer PolicyEnforcer) VerifyNeeded(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor) bool {
	return true
}

func (enforcer PolicyEnforcer) ContinueVerifyOnFailure(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor, partialVerifyResult types.VerifyResult) bool {
	return true
}

func (enforcer PolicyEnforcer) ErrorToVerifyResult(ctx context.Context, subjectRefString string, verifyError error) types.VerifyResult {
	errorReport := verifier.VerifierResult{
		Subject:   subjectRefString,
		IsSuccess: false,
		Results:   []string{fmt.Sprintf("verification failed: %v", verifyError)},
	}

	// TODO  Look for better writing this code segment.
	var reports []interface{}

	reports = append(reports, errorReport)

	return types.VerifyResult{IsSuccess: false, VerifierReports: reports}
}
