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
	"sync"
	"time"

	"github.com/deislabs/ratify/pkg/common"
	e "github.com/deislabs/ratify/pkg/executor"
	"github.com/deislabs/ratify/pkg/executor/config"
	"github.com/deislabs/ratify/pkg/executor/types"
	"github.com/deislabs/ratify/pkg/featureflag"
	"github.com/deislabs/ratify/pkg/metrics"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/policyprovider"
	"github.com/deislabs/ratify/pkg/referrerstore"
	su "github.com/deislabs/ratify/pkg/referrerstore/utils"
	"github.com/deislabs/ratify/pkg/utils"
	vr "github.com/deislabs/ratify/pkg/verifier"
	vt "github.com/deislabs/ratify/pkg/verifier/types"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const (
	defaultVerifyRequestTimeoutMilliseconds = 2900
	defaultMutateRequestTimeoutMilliseconds = 950
)

// Executor describes an execution engine that queries the stores for the supply chain content,
// runs them through the verifiers as governed by the policy enforcer
type Executor struct {
	ReferrerStores []referrerstore.ReferrerStore
	PolicyEnforcer policyprovider.PolicyProvider
	Verifiers      []vr.ReferenceVerifier
	Config         *config.ExecutorConfig
}

// TODO Logging within executor
// VerifySubject verifies the subject and returns results.
func (executor Executor) VerifySubject(ctx context.Context, verifyParameters e.VerifyParameters) (types.VerifyResult, error) {
	if featureflag.UseRegoPolicy.Enabled {
		return executor.verifySubjectForRegoPolicy(ctx, verifyParameters)
	}
	return executor.verifySubjectWithJSONPolicy(ctx, verifyParameters)
}

// verifySubjectWithJSONPolicy verifies the subject with the Json-based policy enforcer.
func (executor Executor) verifySubjectWithJSONPolicy(ctx context.Context, verifyParameters e.VerifyParameters) (types.VerifyResult, error) {
	result, err := executor.verifySubjectInternalWithDecision(ctx, verifyParameters)
	if err != nil {
		// get the result for the error based on the policy.
		// Do we need to consider no referrers as success or failure?
		result = executor.PolicyEnforcer.ErrorToVerifyResult(ctx, verifyParameters.Subject, err)
	}

	return result, nil
}

// verifySubjectForRegoPolicy verifies the subject with results.
// And it also returns the decision based on the verifier results if required.
func (executor Executor) verifySubjectForRegoPolicy(ctx context.Context, verifyParameters e.VerifyParameters) (types.VerifyResult, error) {
	results, err := executor.verifySubjectInternalWithoutDecision(ctx, verifyParameters)
	if err != nil {
		return types.VerifyResult{}, err
	}
	// If it requires embedded Rego Policy Engine make the decision, execute
	// OverallVerifyResult to evaluate the overall result based on the policy.
	result := types.VerifyResult{VerifierReports: results}
	if !featureflag.PassthroughMode.Enabled {
		result.IsSuccess = executor.PolicyEnforcer.OverallVerifyResult(ctx, result.VerifierReports)
	}

	return result, nil
}

// verifySubjectInternalWithDecision verifies the subject and makes the decision
// based on the policy. It is only used internally for the Json-based policy.
func (executor Executor) verifySubjectInternalWithDecision(ctx context.Context, verifyParameters e.VerifyParameters) (types.VerifyResult, error) {
	verifierReports, err := executor.verifySubjectInternalWithoutDecision(ctx, verifyParameters)
	if err != nil {
		return types.VerifyResult{}, err
	}
	if len(verifierReports) == 0 {
		return types.VerifyResult{}, ErrReferrersNotFound
	}

	// Making the decision based on the Json policy.
	overallVerifySuccess := executor.PolicyEnforcer.OverallVerifyResult(ctx, verifierReports)
	return types.VerifyResult{IsSuccess: overallVerifySuccess, VerifierReports: verifierReports}, nil
}

// verifySubjectInternalWithoutDecision verifies the subject and returns result
// without making decisions on the result.
func (executor Executor) verifySubjectInternalWithoutDecision(ctx context.Context, verifyParameters e.VerifyParameters) ([]interface{}, error) {
	subjectReference, err := utils.ParseSubjectReference(verifyParameters.Subject)
	if err != nil {
		return nil, err
	}

	desc, err := su.ResolveSubjectDescriptor(ctx, &executor.ReferrerStores, subjectReference)

	if err != nil {
		return nil, fmt.Errorf("resolving descriptor for the subject failed with error: %w", err)
	}

	logrus.Infof("Resolve of the image completed successfully the digest is %s", desc.Digest)

	subjectReference.Digest = desc.Digest

	verifierReports := make([]interface{}, 0)
	eg, errCtx := errgroup.WithContext(ctx)
	var mu sync.Mutex

	for _, referrerStore := range executor.ReferrerStores {
		referrerStore := referrerStore
		eg.Go(func() error {
			var continuationToken string
			innerGroup, innerErrCtx := errgroup.WithContext(errCtx)
			for {
				referrersResult, err := referrerStore.ListReferrers(errCtx, subjectReference, verifyParameters.ReferenceTypes, continuationToken, desc)
				if err != nil {
					return err
				}
				continuationToken = referrersResult.NextToken
				for _, reference := range referrersResult.Referrers {
					if !executor.PolicyEnforcer.VerifyNeeded(innerErrCtx, subjectReference, reference) {
						continue
					}
					reference := reference
					innerGroup.Go(func() error {
						if featureflag.UseRegoPolicy.Enabled {
							verifyResult, err := executor.verifyReferenceForRegoPolicy(innerErrCtx, subjectReference, reference, referrerStore)
							if err != nil {
								logrus.Errorf("error while verifying reference %+v, err: %v", reference, err)
								return err
							}
							mu.Lock() // locks the verifierReports List for write safety
							defer mu.Unlock()
							verifierReports = append(verifierReports, verifyResult)
						} else {
							verifyResult := executor.verifyReferenceForJSONPolicy(innerErrCtx, subjectReference, reference, referrerStore)
							mu.Lock() // locks the verifierReports List for write safety
							defer mu.Unlock()
							verifierReports = append(verifierReports, verifyResult.VerifierReports...)
						}
						return nil
					})
				}
				if continuationToken == "" {
					break
				}
			}
			return innerGroup.Wait()
		})
	}

	if err = eg.Wait(); err != nil {
		return nil, err
	}

	return verifierReports, nil
}

// verifyReferenceForJSONPolicy verifies the referenced artifact with results
// used for the Json-based policy enforcer.
func (executor Executor) verifyReferenceForJSONPolicy(ctx context.Context, subjectRef common.Reference, referenceDesc ocispecs.ReferenceDescriptor, referrerStore referrerstore.ReferrerStore) types.VerifyResult {
	var verifyResults []interface{}
	var isSuccess = true

	for _, verifier := range executor.Verifiers {
		if verifier.CanVerify(ctx, referenceDesc) {
			verifierStartTime := time.Now()
			verifyResult, err := verifier.Verify(ctx, subjectRef, referenceDesc, referrerStore)
			verifyResult.Subject = subjectRef.String()
			if err != nil {
				verifyResult = vr.VerifierResult{
					IsSuccess: false,
					Name:      verifier.Name(),
					Message:   fmt.Sprintf("an error thrown by the verifier: %v", err)}
			}

			if len(verifier.GetNestedReferences()) > 0 {
				executor.addNestedVerifierResult(ctx, referenceDesc, subjectRef, &verifyResult)
			}

			verifyResult.ArtifactType = referenceDesc.ArtifactType
			verifyResults = append(verifyResults, verifyResult)
			isSuccess = verifyResult.IsSuccess
			metrics.ReportVerifierDuration(ctx, time.Since(verifierStartTime).Milliseconds(), verifier.Name(), subjectRef.String(), isSuccess, err != nil)
			break
		}
	}

	return types.VerifyResult{IsSuccess: isSuccess, VerifierReports: verifyResults}
}

// verifyReferenceForRegoPolicy verifies the referenced artifact with results
// used for Rego-based policy enforcer.
func (executor Executor) verifyReferenceForRegoPolicy(ctx context.Context, subjectRef common.Reference, referenceDesc ocispecs.ReferenceDescriptor, referrerStore referrerstore.ReferrerStore) (types.NestedVerifierReport, error) {
	nestedReport := types.NestedVerifierReport{
		Subject:         subjectRef.String(),
		ArtifactType:    referenceDesc.ArtifactType,
		ReferenceDigest: referenceDesc.Digest.String(),
		VerifierReports: make([]vt.VerifierResult, 0),
		NestedReports:   make([]types.NestedVerifierReport, 0),
	}
	var mu sync.Mutex
	eg, errCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return executor.addNestedReports(errCtx, referenceDesc, subjectRef, &nestedReport)
	})

	for _, verifier := range executor.Verifiers {
		if !verifier.CanVerify(ctx, referenceDesc) {
			continue
		}
		verifier := verifier
		eg.Go(func() error {
			var verifierReport vt.VerifierResult
			verifierStartTime := time.Now()
			verifierResult, err := verifier.Verify(errCtx, subjectRef, referenceDesc, referrerStore)
			if err != nil {
				verifierReport = vt.VerifierResult{
					IsSuccess: false,
					Name:      verifier.Name(),
					Message:   fmt.Sprintf("an error thrown by the verifier: %v", err)}
			} else {
				verifierReport = vt.NewVerifierResult(verifierResult)
			}

			mu.Lock()
			nestedReport.VerifierReports = append(nestedReport.VerifierReports, verifierReport)
			mu.Unlock()

			metrics.ReportVerifierDuration(errCtx, time.Since(verifierStartTime).Milliseconds(), verifier.Name(), subjectRef.String(), verifierReport.IsSuccess, err != nil)
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return types.NestedVerifierReport{}, err
	}
	return nestedReport, nil
}

// addNestedVerifierResult adds the nested verifier result to the parent verify
// result used for Json-based policy enforcer.
func (executor Executor) addNestedVerifierResult(ctx context.Context, referenceDesc ocispecs.ReferenceDescriptor, subjectRef common.Reference, verifyResult *vr.VerifierResult) {
	verifyParameters := e.VerifyParameters{
		Subject:        fmt.Sprintf("%s@%s", subjectRef.Path, referenceDesc.Digest),
		ReferenceTypes: []string{"*"},
	}

	nestedVerifyResult, err := executor.verifySubjectWithJSONPolicy(ctx, verifyParameters)
	if err != nil {
		nestedVerifyResult = executor.PolicyEnforcer.ErrorToVerifyResult(ctx, verifyParameters.Subject, err)
	}

	for _, report := range nestedVerifyResult.VerifierReports {
		if result, ok := report.(vr.VerifierResult); ok {
			verifyResult.NestedResults = append(verifyResult.NestedResults, result)
			if !nestedVerifyResult.IsSuccess {
				verifyResult.IsSuccess = false
				verifyResult.Message = "nested verification failed"
			}
		}
	}
}

// addNestedReports adds the nested verifier reports to the parent report used
// for Rego-based policy enforcer.
func (executor Executor) addNestedReports(ctx context.Context, referenceDes ocispecs.ReferenceDescriptor, subjectRef common.Reference, verifierReport *types.NestedVerifierReport) error {
	verifyParameters := e.VerifyParameters{
		Subject:        fmt.Sprintf("%s@%s", subjectRef.Path, referenceDes.Digest),
		ReferenceTypes: []string{"*"},
	}

	// get nested reports.
	reports, err := executor.verifySubjectForRegoPolicy(ctx, verifyParameters)
	if err != nil {
		return fmt.Errorf("failed to verify nested subject, param: %+v, err: %w", verifyParameters, err)
	}

	// append nested reports to the parent report.
	nestedReports := make([]types.NestedVerifierReport, 0, len(reports.VerifierReports))
	for _, report := range reports.VerifierReports {
		nestedReport, err := types.NewNestedVerifierReport(report)
		if err != nil {
			return err
		}
		nestedReports = append(nestedReports, nestedReport)
	}
	verifierReport.NestedReports = nestedReports
	return nil
}

func (executor Executor) GetVerifyRequestTimeout() time.Duration {
	timeoutMilliSeconds := defaultVerifyRequestTimeoutMilliseconds
	if executor.Config != nil && executor.Config.VerificationRequestTimeout != nil {
		timeoutMilliSeconds = *executor.Config.VerificationRequestTimeout
	}
	return time.Duration(timeoutMilliSeconds) * time.Millisecond
}

func (executor Executor) GetMutationRequestTimeout() time.Duration {
	timeoutMilliSeconds := defaultMutateRequestTimeoutMilliseconds
	if executor.Config != nil && executor.Config.MutationRequestTimeout != nil {
		timeoutMilliSeconds = *executor.Config.MutationRequestTimeout
	}
	return time.Duration(timeoutMilliSeconds) * time.Millisecond
}
