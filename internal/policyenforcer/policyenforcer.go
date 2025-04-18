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

package policyenforcer

import (
	"github.com/ratify-project/ratify-go"
	"github.com/ratify-project/ratify/v2/internal/policyenforcer/factory"
	_ "github.com/ratify-project/ratify/v2/internal/policyenforcer/factory/thresholdpolicy" // Register the threshold policy factory
)

// NewPolicyEnforcer creates a new PolicyEnforcer instance based on the provided options.
func NewPolicyEnforcer(opts *factory.NewPolicyEnforcerOptions) (ratify.PolicyEnforcer, error) {
	return factory.NewPolicyEnforcer(opts)
}
