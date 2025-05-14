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

package thresholdpolicy

import (
	"encoding/json"
	"fmt"

	"github.com/ratify-project/ratify-go"
	"github.com/ratify-project/ratify/v2/internal/policyenforcer/factory"
)

const thresholdPolicyType = "threshold-policy"

// policyRule is defined for rendering [ratify.ThresholdPolicyRule].
type policyRule struct {
	// VerifierName is the name of the verifier to be used for this rule.
	// Optional.
	VerifierName string `json:"verifierName,omitempty"`

	// Threshold is the required number of satisfied nested rules defined in
	// this rule. If not set or set to 0, all nested rules must be satisfied.
	// Optional.
	Threshold int `json:"threshold,omitempty"`

	// Rules hold nested rules that could be applied to referrer artifacts.
	// Optional.
	Rules []*policyRule `json:"rules,omitempty"`
}

type options struct {
	// Policy is the policy rule to be used for this policy enforcer.
	Policy *policyRule `json:"policy"`
}

// init registers the threshold-policy factory via side effects.
// This ensures that the threshold-policy type is available for use in the
// policy enforcer factory.
func init() {
	factory.RegisterPolicyEnforcerFactory(thresholdPolicyType, func(opts *factory.NewPolicyEnforcerOptions) (ratify.PolicyEnforcer, error) {
		raw, err := json.Marshal(opts.Parameters)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal policy options: %w", err)
		}
		var params options
		if err := json.Unmarshal(raw, &params); err != nil {
			return nil, fmt.Errorf("failed to unmarshal policy options: %w", err)
		}

		policy, err := convertPolicy(params.Policy)
		if err != nil {
			return nil, fmt.Errorf("failed to create policy: %w", err)
		}
		return ratify.NewThresholdPolicyEnforcer(policy)
	})
}

// convertPolicy converts a [policyRule] to a [ratify.ThresholdPolicyRule].
func convertPolicy(rule *policyRule) (*ratify.ThresholdPolicyRule, error) {
	if rule == nil {
		return nil, fmt.Errorf("policy options are required")
	}

	rules, err := convertPolicies(rule.Rules)
	if err != nil {
		return nil, fmt.Errorf("failed to create nested rules: %w", err)
	}

	return &ratify.ThresholdPolicyRule{
		Verifier:  rule.VerifierName,
		Threshold: rule.Threshold,
		Rules:     rules,
	}, nil
}

// convertPolicies converts a slice of [policyRule] to a slice of
// [ratify.ThresholdPolicyRule].
func convertPolicies(rules []*policyRule) ([]*ratify.ThresholdPolicyRule, error) {
	if len(rules) == 0 {
		return nil, nil
	}

	policies := make([]*ratify.ThresholdPolicyRule, len(rules))
	for i, rule := range rules {
		policy, err := convertPolicy(rule)
		if err != nil {
			return nil, err
		}
		policies[i] = policy
	}

	return policies, nil
}
