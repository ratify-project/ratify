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

// Convert unversioned StoreStatus to StoreStatus of v1alpha1.
//
//nolint:revive
func Convert_unversioned_StoreStatus_To_v1alpha1_StoreStatus(in *unversioned.StoreStatus, out *StoreStatus, _ conversion.Scope) error {
	return nil
}

// Convert unversioned StoreSpec to VerifieStoreSpecrSpec of v1alpha1.
//
//nolint:revive
func Convert_unversioned_StoreSpec_To_v1alpha1_StoreSpec(spec *unversioned.StoreSpec, out *StoreSpec, _ conversion.Scope) error {
	out.Parameters = spec.Parameters
	return nil
}
