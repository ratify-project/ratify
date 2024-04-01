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
	"github.com/deislabs/ratify/internal/constants"
	vr "github.com/deislabs/ratify/pkg/verifier"
)

type ActiveVerifiers struct {
	NamespacedVerifiers map[string]map[string]vr.ReferenceVerifier
}

func NewActiveVerifiers() ActiveVerifiers {
	return ActiveVerifiers{
		NamespacedVerifiers: make(map[string]map[string]vr.ReferenceVerifier),
	}
}

// GetVerifiers implements the Verifiers interface.
// It returns a list of verifiers for the given scope. If no verifiers are found for the given scope, it returns cluster-wide verifiers.
func (v *ActiveVerifiers) GetVerifiers(scope string) []vr.ReferenceVerifier {
	verifiers := []vr.ReferenceVerifier{}
	for _, verifier := range v.NamespacedVerifiers[scope] {
		verifiers = append(verifiers, verifier)
	}

	if len(verifiers) == 0 && scope != constants.EmptyNamespace {
		for _, verifier := range v.NamespacedVerifiers[constants.EmptyNamespace] {
			verifiers = append(verifiers, verifier)
		}
	}
	return verifiers
}

func (v *ActiveVerifiers) AddVerifier(scope, verifierName string, verifier vr.ReferenceVerifier) {
	if _, ok := v.NamespacedVerifiers[scope]; !ok {
		v.NamespacedVerifiers[scope] = make(map[string]vr.ReferenceVerifier)
	}
	v.NamespacedVerifiers[scope][verifierName] = verifier
}

func (v *ActiveVerifiers) DeleteVerifier(scope, verifierName string) {
	if verifiers, ok := v.NamespacedVerifiers[scope]; ok {
		delete(verifiers, verifierName)
		if len(verifiers) == 0 {
			delete(v.NamespacedVerifiers, scope)
		}
	}
}

func (v *ActiveVerifiers) IsEmpty() bool {
	return len(v.NamespacedVerifiers) == 0
}

func (v *ActiveVerifiers) GetVerifierCount() int {
	count := 0
	for _, verifiers := range v.NamespacedVerifiers {
		count += len(verifiers)
	}
	return count
}
