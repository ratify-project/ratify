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
	"testing"

	"github.com/deislabs/ratify/pkg/common"
	vt "github.com/deislabs/ratify/pkg/executor/types"
	"github.com/deislabs/ratify/pkg/ocispecs"
	pc "github.com/deislabs/ratify/pkg/policyprovider/config"
	pf "github.com/deislabs/ratify/pkg/policyprovider/factory"
	"github.com/deislabs/ratify/pkg/policyprovider/types"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
)

func TestPolicyEnforcer_ContinueVerifyOnFailure(t *testing.T) {
	configPolicyConfig := map[string]interface{}{
		"name": "configPolicy",
		"artifactVerificationPolicies": map[string]types.ArtifactTypeVerifyPolicy{
			"application/vnd.cncf.notary.v2": "any",
			"org.example.sbom.v0":            "all",
			"default":                        "any",
		},
	}
	config := pc.PoliciesConfig{
		Version:      "1.0.0",
		PolicyPlugin: configPolicyConfig,
	}

	policyEnforcer, err := pf.CreatePolicyProviderFromConfig(config)

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

// TODO: add test for OverallVerifyResult
