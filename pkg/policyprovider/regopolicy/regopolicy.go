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

package regopolicy

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
	"github.com/deislabs/ratify/pkg/policyprovider/policyevaluation"
	"github.com/sirupsen/logrus"
)

const (
	nestedReportsField = "nested_reports"
)

type policyEnforcer struct {
	Policy    string
	OpaEngine policyevaluation.PolicyEvaluator
}

type policyEnforcerConf struct {
	Name   string `json:"name"`
	Policy string `json:"policy"`
}

type RegoPolicyFactory struct{}

// init calls Register for our rego policy provider.
func init() {
	pf.Register("regoPolicy", &RegoPolicyFactory{})
}

// Create creates a new policy enforcer based on the policy provided in config.
func (f *RegoPolicyFactory) Create(policyConfig config.PolicyPluginConfig) (policyprovider.PolicyProvider, error) {
	policyEnforcer := &policyEnforcer{}

	conf := policyEnforcerConf{}
	policyProviderConfigBytes, err := json.Marshal(policyConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal policy config: %w", err)
	}

	if err := json.Unmarshal(policyProviderConfigBytes, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse policy provider configuration: %w", err)
	}
	if conf.Policy == "" {
		return nil, fmt.Errorf("policy is required for rego policy provider")
	}

	engine, err := policyevaluation.NewOpaEngine(conf.Policy)
	if err != nil {
		return nil, fmt.Errorf("failed to create OPA engine: %w", err)
	}

	policyEnforcer.Policy = conf.Policy
	policyEnforcer.OpaEngine = engine

	return policyEnforcer, nil
}

// VerifyNeeded determines if verification should be performed for a given artifact.
func (e *policyEnforcer) VerifyNeeded(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor) bool {
	return true
}

// ContinueVerifyOnFailure determines if verification should continue if a previous verification failed.
func (e *policyEnforcer) ContinueVerifyOnFailure(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor, partialVerifyResult types.VerifyResult) bool {
	return true
}

// ErrorToVerifyResult converts an error to a VerifyResult.
func (enforcer *policyEnforcer) ErrorToVerifyResult(ctx context.Context, subjectRefString string, verifyError error) types.VerifyResult {
	return types.VerifyResult{}
}

// OverallVerifyResult determines if the overall verification result should be a success or failure.
func (enforcer *policyEnforcer) OverallVerifyResult(ctx context.Context, verifierReports []interface{}) bool {
	if len(verifierReports) == 0 {
		logrus.Errorf("no verifier reports to evaluate")
		return false
	}

	nestedReports := map[string]interface{}{}
	nestedReports[nestedReportsField] = verifierReports
	result, err := enforcer.OpaEngine.Evaluate(ctx, nestedReports)
	if err != nil {
		logrus.Errorf("failed to evaluate policy: %v", err)
		return false
	}
	return result
}
