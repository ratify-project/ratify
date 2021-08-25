package core

import (
	"context"
	"fmt"

	"github.com/notaryproject/hora/pkg/common"
	e "github.com/notaryproject/hora/pkg/executor"
	"github.com/notaryproject/hora/pkg/executor/types"
	"github.com/notaryproject/hora/pkg/ocispecs"
	"github.com/notaryproject/hora/pkg/policyprovider"
	"github.com/notaryproject/hora/pkg/referrerstore"
	"github.com/notaryproject/hora/pkg/utils"
	vr "github.com/notaryproject/hora/pkg/verifier"
)

type Executor struct {
	ReferrerStores []referrerstore.ReferrerStore
	PolicyEnforcer policyprovider.PolicyProvider
	Verifiers      []vr.ReferenceVerifier
}

// TODO Logging within executor
func (executor Executor) VerifySubject(ctx context.Context, verifyParameters e.VerifyParameters) (types.VerifyResult, error) {
	result, err := executor.verifySubjectInternal(ctx, verifyParameters)

	if err != nil {
		// get the result for the error based on the policy.
		// Do we need to consider no referrers as success or failure?
		result = executor.PolicyEnforcer.ErrorToVerifyResult(ctx, verifyParameters.Subject, err)
	}

	return result, nil
}

func (executor Executor) verifySubjectInternal(ctx context.Context, verifyParameters e.VerifyParameters) (types.VerifyResult, error) {
	subjectReference, err := utils.ParseSubjectReference(verifyParameters.Subject)
	if err != nil {
		return types.VerifyResult{}, err
	}

	var verifierReports []interface{}

	for _, referrerStore := range executor.ReferrerStores {
		var continuationToken string
		for {
			referrersResult, err := referrerStore.ListReferrers(ctx, subjectReference, verifyParameters.ReferenceTypes, continuationToken)

			if err != nil {
				return types.VerifyResult{}, err
			}

			continuationToken = referrersResult.NextToken

			for _, reference := range referrersResult.Referrers {

				if executor.PolicyEnforcer.VerifyNeeded(ctx, subjectReference, reference) {
					verifyResult := executor.verifyReference(ctx, subjectReference, reference, referrerStore)
					verifierReports = append(verifierReports, verifyResult.VerifierReports...)

					if !verifyResult.IsSuccess {
						result := types.VerifyResult{IsSuccess: false, VerifierReports: verifierReports}
						if !executor.PolicyEnforcer.ContinueVerifyOnFailure(ctx, subjectReference, reference, result) {
							return result, nil
						}
					}
				}
			}

			if continuationToken == "" {
				break
			}
		}

	}

	if len(verifierReports) == 0 {
		return types.VerifyResult{}, ReferrersNotFound
	}

	return types.VerifyResult{IsSuccess: true, VerifierReports: verifierReports}, nil
}

func (ex Executor) verifyReference(ctx context.Context, subjectRef common.Reference, referenceDesc ocispecs.ReferenceDescriptor, referrerStore referrerstore.ReferrerStore) types.VerifyResult {
	var verifyResults []interface{}
	var isSuccess = true

	for _, verifier := range ex.Verifiers {
		if verifier.CanVerify(ctx, referenceDesc) {
			verifyResult, err := verifier.Verify(ctx, subjectRef, referenceDesc, referrerStore, ex)
			verifyResult.Subject = subjectRef.String()
			if err != nil {
				verifyResult = vr.VerifierResult{
					IsSuccess: false,
					Name:      verifier.Name(),
					Results:   []string{fmt.Sprintf("an error thrown by the verifier %v", err)}}
			}

			verifyResults = append(verifyResults, verifyResult)

			isSuccess = verifyResult.IsSuccess
			break
		}

	}

	return types.VerifyResult{IsSuccess: isSuccess, VerifierReports: verifyResults}
}
