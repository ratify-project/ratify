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

package verifiers

import (
	vr "github.com/deislabs/ratify/pkg/verifier"
)

// ActiveVerifiers implements VerifierManger interface.
type ActiveVerifiers struct {
	// TODO: Implement concurrent safety using sync.Map
	// The structure of the map is as follows:
	// The first level maps from scope to verifiers
	// The second level maps from verifier name to verifier
	// Example:
	// {
	//   "namespace1": {
	//     "verifier1": verifier1,
	//     "verifier2": verifier2
	//   }
	// }
	// Note: Scope is utilized for organizing and isolating verifiers. In a Kubernetes (K8s) environment, the scope can be either a namespace or an empty string ("") for cluster-wide verifiers.
	ScopedVerifiers map[string]map[string]vr.ReferenceVerifier
}

func NewActiveVerifiers() VerifierManager {
	return &ActiveVerifiers{
		ScopedVerifiers: make(map[string]map[string]vr.ReferenceVerifier),
	}
}

// GetVerifiers implements the Verifiers interface.
// It returns a list of verifiers for the given scope. If no verifiers are found for the given scope, it returns cluster-wide verifiers.
// TODO: Current implementation fetches verifiers for all namespaces including cluster-wide ones. Will support actual namespace based verifiers in future.
func (v *ActiveVerifiers) GetVerifiers(_ string) []vr.ReferenceVerifier {
	verifiers := []vr.ReferenceVerifier{}
	for _, namespacedVerifiers := range v.ScopedVerifiers {
		for _, verifier := range namespacedVerifiers {
			verifiers = append(verifiers, verifier)
		}
	}
	return verifiers
}

func (v *ActiveVerifiers) AddVerifier(scope, verifierName string, verifier vr.ReferenceVerifier) {
	if _, ok := v.ScopedVerifiers[scope]; !ok {
		v.ScopedVerifiers[scope] = make(map[string]vr.ReferenceVerifier)
	}
	v.ScopedVerifiers[scope][verifierName] = verifier
}

func (v *ActiveVerifiers) DeleteVerifier(scope, verifierName string) {
	if verifiers, ok := v.ScopedVerifiers[scope]; ok {
		delete(verifiers, verifierName)
	}
}

func (v *ActiveVerifiers) IsEmpty() bool {
	return v.GetVerifierCount() == 0
}

func (v *ActiveVerifiers) GetVerifierCount() int {
	count := 0
	for _, verifiers := range v.ScopedVerifiers {
		count += len(verifiers)
	}
	return count
}
