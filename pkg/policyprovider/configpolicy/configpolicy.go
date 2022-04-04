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

package config

import (
	"context"
	"fmt"

	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/executor/types"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/policyprovider"
	"github.com/deislabs/ratify/pkg/policyprovider/config"
	vt "github.com/deislabs/ratify/pkg/policyprovider/types"
	"github.com/deislabs/ratify/pkg/verifier"
)

// PolicyEnforcer describes different polices that are enforced during verification
type PolicyEnforcer struct {
	ArtifactTypePolicies map[string]vt.ArtifactTypeVerifyPolicy
}

const defaultPolicyName string = "default"

// CreatePolicyEnforcerFromConfig creates a policy enforcer from the given configuration
func CreatePolicyEnforcerFromConfig(policiesConfig config.PoliciesConfig) (policyprovider.PolicyProvider, error) {
	policyEnforcer := PolicyEnforcer{}
	if policiesConfig.ArtifactVerificationPolicies == nil {
		policyEnforcer.ArtifactTypePolicies = map[string]vt.ArtifactTypeVerifyPolicy{}
	} else {
		policyEnforcer.ArtifactTypePolicies = policiesConfig.ArtifactVerificationPolicies
	}
	if policyEnforcer.ArtifactTypePolicies[defaultPolicyName] == "" {
		policyEnforcer.ArtifactTypePolicies[defaultPolicyName] = vt.AnyVerifySuccess
	}
	return &policyEnforcer, nil
}

func (enforcer PolicyEnforcer) VerifyNeeded(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor) bool {
	return true
}

func (enforcer PolicyEnforcer) ContinueVerifyOnFailure(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor, partialVerifyResult types.VerifyResult) bool {
	artifactType := referenceDesc.ArtifactType
	policy := enforcer.ArtifactTypePolicies[artifactType]
	if policy == "" {
		policy = enforcer.ArtifactTypePolicies[defaultPolicyName]
	}
	if policy == vt.AnyVerifySuccess {
		return true
	} else {
		return false
	}
}

func (enforcer PolicyEnforcer) ErrorToVerifyResult(ctx context.Context, subjectRefString string, verifyError error) types.VerifyResult {
	errorReport := verifier.VerifierResult{
		Subject:   subjectRefString,
		IsSuccess: false,
		Results:   []string{fmt.Sprintf("verification failed: %v", verifyError)},
	}
	var reports []interface{}
	reports = append(reports, errorReport)
	return types.VerifyResult{IsSuccess: false, VerifierReports: reports}
}

func (enforcer PolicyEnforcer) OverallVerifyResult(ctx context.Context, verifierReports []interface{}) bool {
	// use boolean map to track if each artifact type policy constraint is satisfied
	verifySuccess := map[string]bool{}
	for artifactType := range enforcer.ArtifactTypePolicies {
		// add all policies excpept for default
		if artifactType != defaultPolicyName {
			verifySuccess[artifactType] = false
		}
	}

	for _, report := range verifierReports {
		castedReport := report.(verifier.VerifierResult)
		// the artifact type of the verified artifact matches type specified in policy
		if policyType, ok := enforcer.ArtifactTypePolicies[castedReport.ArtifactType]; ok {
			// the artifact type of the verified artifact matches type specified in policy
			if policyType == vt.AnyVerifySuccess && castedReport.IsSuccess {
				// if policy is 'any' and report is successful
				verifySuccess[castedReport.ArtifactType] = true
			} else if policyType == vt.AllVerifySuccess {
				// if policy is 'all'
				if !castedReport.IsSuccess {
					// return false after first failure
					return false
				}
				verifySuccess[castedReport.ArtifactType] = true
			}
		}
	}

	// all booleans in map must be true for overall success to be true
	for artifactType := range verifySuccess {
		if !verifySuccess[artifactType] {
			return false
		}
	}
	return true
}
