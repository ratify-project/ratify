package config

import (
	"context"
	"fmt"

	"github.com/deislabs/hora/pkg/common"
	"github.com/deislabs/hora/pkg/executor/types"
	"github.com/deislabs/hora/pkg/ocispecs"
	"github.com/deislabs/hora/pkg/policyprovider"
	"github.com/deislabs/hora/pkg/policyprovider/config"
	vt "github.com/deislabs/hora/pkg/policyprovider/types"
	"github.com/deislabs/hora/pkg/verifier"
)

type PolicyEnforcer struct {
	ArtifactTypePolicies map[string]vt.ArtifactTypeVerifyPolicy
}

func CreatePolicyEnforcerFromConfig(policiesConfig config.PoliciesConfig) (policyprovider.PolicyProvider, error) {
	policyEnforcer := PolicyEnforcer{}
	if policiesConfig.ArtifactVerificationPolicies == nil {
		policyEnforcer.ArtifactTypePolicies = map[string]vt.ArtifactTypeVerifyPolicy{}
	} else {
		policyEnforcer.ArtifactTypePolicies = policiesConfig.ArtifactVerificationPolicies
	}
	if policyEnforcer.ArtifactTypePolicies["default"] == "" {
		policyEnforcer.ArtifactTypePolicies["default"] = vt.AnyVerifySuccess
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
		policy = enforcer.ArtifactTypePolicies["default"]
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

	// TODO  Look for better writing this code segment.
	var reports []interface{}

	reports = append(reports, errorReport)

	return types.VerifyResult{IsSuccess: false, VerifierReports: reports}
}
