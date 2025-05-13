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

package factory

import (
	"fmt"

	"github.com/ratify-project/ratify-go"
)

// NewPolicyEnforcerOptions is the options for creating a new PolicyEnforcer.
type NewPolicyEnforcerOptions struct {
	// Type represents a specific implementation of a policy enforcer. Required.
	Type string `json:"type"`

	// Parameters is additional parameters for the policy enforcer. Optional.
	Parameters any `json:"parameters,omitempty"`
}

// registeredPolicyEnforcers saves the registered policy enforcer factories.
var registeredPolicyEnforcers = make(map[string]func(*NewPolicyEnforcerOptions) (ratify.PolicyEnforcer, error))

// RegisterPolicyEnforcer registers a policy enforcer factory to the system.
func RegisterPolicyEnforcerFactory(policyType string, create func(*NewPolicyEnforcerOptions) (ratify.PolicyEnforcer, error)) {
	if policyType == "" {
		panic("policy type cannot be empty")
	}
	if create == nil {
		panic("policy factory cannot be nil")
	}
	if _, registered := registeredPolicyEnforcers[policyType]; registered {
		panic("policy factory already registered")
	}
	registeredPolicyEnforcers[policyType] = create
}

// NewPolicyEnforcer creates a new PolicyEnforcer instance based on the provided options.
func NewPolicyEnforcer(opts *NewPolicyEnforcerOptions) (ratify.PolicyEnforcer, error) {
	if opts == nil {
		return nil, nil
	}
	if opts.Type == "" {
		return nil, fmt.Errorf("policy type is not provided in the policy options")
	}
	policyFactory, ok := registeredPolicyEnforcers[opts.Type]
	if !ok {
		return nil, fmt.Errorf("policy factory of type %s is not registered", opts.Type)
	}
	return policyFactory(opts)
}
