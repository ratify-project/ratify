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

package core

import (
	"context"
	"fmt"

	"github.com/deislabs/ratify/pkg/common"
	e "github.com/deislabs/ratify/pkg/executor"
	"github.com/deislabs/ratify/pkg/executor/types"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/policyprovider"
	"github.com/deislabs/ratify/pkg/referrerstore"
	"github.com/deislabs/ratify/pkg/utils"
	vr "github.com/deislabs/ratify/pkg/verifier"
)

// Executor describes an execution engine that queries the stores for the supply chain content,
// runs them through the verifiers as governed by the policy enforcer
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
	anyVerifySuccess := map[string]bool{}
	for _, referenceType := range verifyParameters.ReferenceTypes {
		anyVerifySuccess[referenceType] = false
	}

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
					} else {
						anyVerifySuccess[reference.ArtifactType] = true
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

	overallVerifySuccess := true
	for _, referenceType := range verifyParameters.ReferenceTypes {
		if anyVerifySuccess[referenceType] == false {
			overallVerifySuccess = false
			break
		}
	}

	return types.VerifyResult{IsSuccess: overallVerifySuccess, VerifierReports: verifierReports}, nil
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
