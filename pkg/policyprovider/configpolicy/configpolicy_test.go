package config

import (
	"context"
	"github.com/deislabs/hora/pkg/common"
	vt "github.com/deislabs/hora/pkg/executor/types"
	"github.com/deislabs/hora/pkg/ocispecs"
	pc "github.com/deislabs/hora/pkg/policyprovider/config"
	"github.com/deislabs/hora/pkg/policyprovider/types"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
	"testing"
)

func TestPolicyEnforcer_ContinueVerifyOnFailure(t *testing.T) {
	config := pc.PoliciesConfig{
		Version: "1.0.0",
		ArtifactVerificationPolicies: map[string]types.ArtifactTypeVerifyPolicy{
			"application/vnd.cncf.notary.v2": "any",
			"org.example.sbom.v0":            "all",
			"default":                        "any",
		},
	}

	policyEnforcer, err := CreatePolicyEnforcerFromConfig(config)

	if err != nil {
		t.Fatalf("PolicyEnforcer should create from PoliciesConfig")
	}

	ctx := context.Background()
	subjectReference := common.Reference{
		Path:     "",
		Digest:   "",
		Tag:      "",
		Original: "",
	}
	referenceDesc := ocispecs.ReferenceDescriptor{
		Descriptor:   oci.Descriptor{},
		ArtifactType: "application/vnd.cncf.notary.v2",
	}
	result := vt.VerifyResult{
		IsSuccess:       false,
		VerifierReports: nil,
	}

	check := policyEnforcer.ContinueVerifyOnFailure(ctx, subjectReference, referenceDesc, result)

	if check != true {
		t.Fatalf("For policy of 'any' PolicyEnforcer should allow continuing on verify failure")
	}

	referenceDesc.ArtifactType = "org.example.sbom.v0"

	check = policyEnforcer.ContinueVerifyOnFailure(ctx, subjectReference, referenceDesc, result)

	if check != false {
		t.Fatalf("For policy 'all' PolicyEnforcer should not allow continuing on verify failure")
	}

	referenceDesc.ArtifactType = "unknown"

	check = policyEnforcer.ContinueVerifyOnFailure(ctx, subjectReference, referenceDesc, result)

	if check != true {
		t.Fatalf("For artifact types without a policy the default policy should be followed")
	}
}
