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

package policies

import (
	"github.com/deislabs/ratify/internal/constants"
	"github.com/deislabs/ratify/pkg/policyprovider"
)

// PolicyWrapper wraps policy provider with its policy name.
type PolicyWrapper struct {
	Name   string
	Policy policyprovider.PolicyProvider
}

// ActivePolicies implements PolicyManager interface.
type ActivePolicies struct {
	// TODO: Implement concurrent safety using sync.Map
	// ScopedPolicies is a mapping from scope to a policy.
	// Note: Scope is utilized for organizing and isolating verifiers. In a Kubernetes (K8s) environment, the scope can be either a namespace or an empty string ("") for cluster-wide verifiers.
	ScopedPolicies map[string]PolicyWrapper
}

func NewActivePolicies() PolicyManager {
	return &ActivePolicies{
		ScopedPolicies: make(map[string]PolicyWrapper),
	}
}

// GetPolicy fulfills the PolicyManager interface.
// It returns the policy for the given scope. If no policy is found for the given scope, it returns cluster-wide policy.
// TODO: Current implementation always fetches the cluster-wide policy. Will implement the logic to fetch the policy for the given scope.
func (p *ActivePolicies) GetPolicy(_ string) policyprovider.PolicyProvider {
	policy, ok := p.ScopedPolicies[constants.EmptyNamespace]
	if ok {
		return policy.Policy
	}
	return nil
}

// AddPolicy fulfills the PolicyManager interface.
// It adds the given policy under the given scope.
func (p *ActivePolicies) AddPolicy(scope, policyName string, policy policyprovider.PolicyProvider) {
	p.ScopedPolicies[scope] = PolicyWrapper{
		Name:   policyName,
		Policy: policy,
	}
}

// DeletePolicy fulfills the PolicyManager interface.
// It deletes the policy from the given scope.
func (p *ActivePolicies) DeletePolicy(scope, policyName string) {
	if policy, ok := p.ScopedPolicies[scope]; ok {
		if policy.Name == policyName {
			delete(p.ScopedPolicies, scope)
		}
	}
}

// IsEmpty fulfills the PolicyManager interface.
// IsEmpty returns true if there are no policies.
func (p *ActivePolicies) IsEmpty() bool {
	return len(p.ScopedPolicies) == 0
}
