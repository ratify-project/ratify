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

type policyWrapper struct {
	Name   string
	Policy policyprovider.PolicyProvider
}

type ActivePolicies struct {
	NamespacedPolicies map[string]policyWrapper
}

func NewActivePolicies() ActivePolicies {
	return ActivePolicies{
		NamespacedPolicies: make(map[string]policyWrapper),
	}
}

// GetPolicy implements the Policies interface.
// It returns the policy for the given scope. If no policy is found for the given scope, it returns cluster-wide policy.
func (p *ActivePolicies) GetPolicy(scope string) policyprovider.PolicyProvider {
	if policy, ok := p.NamespacedPolicies[scope]; ok {
		return policy.Policy
	}
	if scope != constants.EmptyNamespace {
		if policy, ok := p.NamespacedPolicies[constants.EmptyNamespace]; ok {
			return policy.Policy
		}
	}

	return nil
}

func (p *ActivePolicies) AddPolicy(scope, policyName string, policy policyprovider.PolicyProvider) {
	p.NamespacedPolicies[scope] = policyWrapper{
		Name:   policyName,
		Policy: policy,
	}
}

func (p *ActivePolicies) DeletePolicy(scope, policyName string) {
	if policy, ok := p.NamespacedPolicies[scope]; ok {
		if policy.Name == policyName {
			delete(p.NamespacedPolicies, scope)
		}
	}
}

func (p *ActivePolicies) IsEmpty() bool {
	return len(p.NamespacedPolicies) == 0
}
