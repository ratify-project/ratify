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

package policyengine

import "context"

// PolicyEngine is an interface that represents a policy engine.
type PolicyEngine interface {
	// Evaluate evaluates the policy with the given input.
	// input is the verifier reports that engine evaluates against.
	// result indicates whether the input satisfies the policy.
	// err indicates an error happened during the evaluation.
	Evaluate(ctx context.Context, input map[string]interface{}) (result bool, err error)
}
