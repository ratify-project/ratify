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

package v1alpha1

import (
	unversioned "github.com/ratify-project/ratify/api/unversioned"
	conversion "k8s.io/apimachinery/pkg/conversion"
)

// Convert unversioned VerifierStatus to VerifierStatus of v1alpha1.
//
//nolint:revive
func Convert_unversioned_VerifierStatus_To_v1alpha1_VerifierStatus(status *unversioned.VerifierStatus, out *VerifierStatus, _ conversion.Scope) error {
	return nil
}

// Convert unversioned VerifierSpec to VerifierSpec of v1alpha1.
//
//nolint:revive
func Convert_unversioned_VerifierSpec_To_v1alpha1_VerifierSpec(spec *unversioned.VerifierSpec, out *VerifierSpec, _ conversion.Scope) error {
	out.Parameters = spec.Parameters
	return nil
}
