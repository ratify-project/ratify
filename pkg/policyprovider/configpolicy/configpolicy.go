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

package configpolicy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/executor/types"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/policyprovider"
	"github.com/deislabs/ratify/pkg/policyprovider/config"
	pf "github.com/deislabs/ratify/pkg/policyprovider/factory"
	vt "github.com/deislabs/ratify/pkg/policyprovider/types"
	"github.com/deislabs/ratify/pkg/verifier"
)

// PolicyEnforcer describes different polices that are enforced during verification
type PolicyEnforcer struct {
	ArtifactTypePolicies map[string]vt.ArtifactTypeVerifyPolicy
}

type configPolicyEnforcerConf struct {
	Name                         string                                 `json:"name"`
	ArtifactVerificationPolicies map[string]vt.ArtifactTypeVerifyPolicy `json:"artifactVerificationPolicies,omitempty"`
}

const defaultPolicyName string = "default"

type configPolicyFactory struct{}

// init calls Register for our config policy provider
func init() {
	pf.Register("configPolicy", &configPolicyFactory{})
}

// Create initializes a new policy provider based on the provider selected in config
func (f *configPolicyFactory) Create(policyConfig config.PolicyPluginConfig) (policyprovider.PolicyProvider, error) {
	policyEnforcer := PolicyEnforcer{}

	conf := configPolicyEnforcerConf{}
	policyProviderConfigBytes, err := json.Marshal(policyConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal policy config: %v", err)
	}

	if err := json.Unmarshal(policyProviderConfigBytes, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse policy provider configuration: %v", err)
	}

	if conf.ArtifactVerificationPolicies == nil {
		policyEnforcer.ArtifactTypePolicies = map[string]vt.ArtifactTypeVerifyPolicy{}
	} else {
		policyEnforcer.ArtifactTypePolicies = conf.ArtifactVerificationPolicies
	}
	if policyEnforcer.ArtifactTypePolicies[defaultPolicyName] == "" {
		policyEnforcer.ArtifactTypePolicies[defaultPolicyName] = vt.AllVerifySuccess
	}
	return &policyEnforcer, nil
}

// VerifyNeeded determines if the given subject/reference artifact should be verified
func (enforcer PolicyEnforcer) VerifyNeeded(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor) bool {
	return true
}

// ContinueVerifyOnFailure determines if the given error can be ignored and verification can be continued.
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

// ErrorToVerifyResult converts an error to a properly formatted verify result
func (enforcer PolicyEnforcer) ErrorToVerifyResult(ctx context.Context, subjectRefString string, verifyError error) types.VerifyResult {
	errorReport := verifier.VerifierResult{
		Subject:   subjectRefString,
		IsSuccess: false,
		Message:   fmt.Sprintf("verification failed: %v", verifyError),
	}
	var reports []interface{}
	reports = append(reports, errorReport)
	return types.VerifyResult{IsSuccess: false, VerifierReports: reports}
}

// OverallVerifyResult determines the final outcome of verification that is constructed using the results from
// individual verifications
func (enforcer PolicyEnforcer) OverallVerifyResult(ctx context.Context, verifierReports []interface{}) bool {
	if len(verifierReports) <= 0 {
		return false
	}

	// use boolean map to track if each artifact type policy constraint is satisfied
	verifySuccess := map[string]bool{}
	for artifactType := range enforcer.ArtifactTypePolicies {
		// add all policies except for default
		if artifactType != defaultPolicyName {
			verifySuccess[artifactType] = false
		}
	}

	for _, report := range verifierReports {
		castedReport := report.(verifier.VerifierResult)
		// extract the policy for the artifact type of the verified artifact if specified
		policyType, ok := enforcer.ArtifactTypePolicies[castedReport.ArtifactType]
		// if artifact type policy not specified, set policy to be default policy and add artifact type to success map
		if !ok {
			policyType = enforcer.ArtifactTypePolicies[defaultPolicyName]
			// set the artifact type success field in map to false to start
			verifySuccess[castedReport.ArtifactType] = false
			// add the unspecified artifact type to the enforcer's artifact type map
			enforcer.ArtifactTypePolicies[castedReport.ArtifactType] = policyType
		}

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

	// all booleans in map must be true for overall success to be true
	for artifactType := range verifySuccess {
		if !verifySuccess[artifactType] {
			return false
		}
	}
	return true
}
