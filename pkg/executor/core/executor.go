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
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/policyprovider"
	"github.com/deislabs/ratify/pkg/referrerstore"
	su "github.com/deislabs/ratify/pkg/referrerstore/utils"
	"github.com/deislabs/ratify/pkg/utils"
	vr "github.com/deislabs/ratify/pkg/verifier"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const defaultRequestTimeoutMilliseconds = 2800

// Executor describes an execution engine that queries the stores for the supply chain content,
// runs them through the verifiers as governed by the policy enforcer
type Executor struct {
	ReferrerStores []referrerstore.ReferrerStore
	PolicyEnforcer policyprovider.PolicyProvider
	Verifiers      []vr.ReferenceVerifier
	Config         *config.ExecutorConfig
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

func (executor Executor) GetVerifyRequestTimeout() time.Duration {
	timeoutMilliSeconds := defaultRequestTimeoutMilliseconds
	if executor.Config != nil && executor.Config.RequestTimeout != nil {
		timeoutMilliSeconds = *executor.Config.RequestTimeout
	}
	return time.Duration(timeoutMilliSeconds) * time.Millisecond
}

func (executor Executor) verifySubjectInternal(ctx context.Context, verifyParameters e.VerifyParameters) (types.VerifyResult, error) {
	subjectReference, err := utils.ParseSubjectReference(verifyParameters.Subject)
	if err != nil {
		return types.VerifyResult{}, err
	}

	desc, err := su.ResolveSubjectDescriptor(ctx, &executor.ReferrerStores, subjectReference)

	if err != nil {
		return types.VerifyResult{}, fmt.Errorf("resolving descriptor for the subject failed with error %v", err)
	}

	logrus.Infof("Resolve of the image completed successfully the digest is %s", desc.Digest)

	subjectReference.Digest = desc.Digest

	var verifierReports []interface{}
	eg, errCtx := errgroup.WithContext(ctx)
	var mu sync.Mutex

	for _, referrerStore := range executor.ReferrerStores {
		referrerStore := referrerStore
		eg.Go(func() error {
			var continuationToken string
			wg := sync.WaitGroup{}
			for {
				referrersResult, err := referrerStore.ListReferrers(errCtx, subjectReference, verifyParameters.ReferenceTypes, continuationToken, desc)
				if err != nil {
					return err
				}
				continuationToken = referrersResult.NextToken
				for _, reference := range referrersResult.Referrers {
					wg.Add(1)
					go func(reference ocispecs.ReferenceDescriptor) {
						defer wg.Done()
						if executor.PolicyEnforcer.VerifyNeeded(ctx, subjectReference, reference) {
							verifyResult := executor.verifyReference(errCtx, subjectReference, desc, reference, referrerStore)
							mu.Lock() // locks the verifierReports List for write safety
							defer mu.Unlock()
							verifierReports = append(verifierReports, verifyResult.VerifierReports...)
						}
					}(reference)
				}
				if continuationToken == "" {
					break
				}
			}
			wg.Wait()
			return nil
		})
	}

	if err = eg.Wait(); err != nil {
		return types.VerifyResult{}, err
	}

	if len(verifierReports) == 0 {
		return types.VerifyResult{}, ErrReferrersNotFound
	}

	overallVerifySuccess := executor.PolicyEnforcer.OverallVerifyResult(ctx, verifierReports)
	return types.VerifyResult{IsSuccess: overallVerifySuccess, VerifierReports: verifierReports}, nil
}

func (ex Executor) verifyReference(ctx context.Context, subjectRef common.Reference, subjectDesc *ocispecs.SubjectDescriptor, referenceDesc ocispecs.ReferenceDescriptor, referrerStore referrerstore.ReferrerStore) types.VerifyResult {
	var verifyResults []interface{}
	var isSuccess = true

	for _, verifier := range ex.Verifiers {

		if verifier.CanVerify(ctx, referenceDesc) {
			verifyResult, err := verifier.Verify(ctx, subjectRef, referenceDesc, referrerStore)
			verifyResult.Subject = subjectRef.String()
			if err != nil {
				verifyResult = vr.VerifierResult{
					IsSuccess: false,
					Name:      verifier.Name(),
					Message:   fmt.Sprintf("an error thrown by the verifier: %v", err)}
			}

			if shouldCheckNested, ok := ex.Config.NestedReferences[referenceDesc.ArtifactType]; ok && shouldCheckNested {
				ex.addNestedVerifierResult(ctx, referenceDesc, subjectRef, &verifyResult)
			}

			verifyResult.ArtifactType = referenceDesc.ArtifactType
			verifyResults = append(verifyResults, verifyResult)
			isSuccess = verifyResult.IsSuccess
			break
		}
	}

	return types.VerifyResult{IsSuccess: isSuccess, VerifierReports: verifyResults}
}

func (ex Executor) addNestedVerifierResult(ctx context.Context, referenceDesc ocispecs.ReferenceDescriptor, subjectRef common.Reference, verifyResult *vr.VerifierResult) {
	verifyParameters := e.VerifyParameters{
		Subject:        fmt.Sprintf("%s@%s", subjectRef.Path, referenceDesc.Digest),
		ReferenceTypes: []string{"*"},
	}

	// error is never returned from ex.VerifySubject, if an err occurs, the result contains the error message
	nestedVerifyResult, err := ex.VerifySubject(ctx, verifyParameters)

	// panic here because ex.VerifySubject doesn't currently return an error
	if err != nil {
		panic(err)
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
