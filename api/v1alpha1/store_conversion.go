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
	unsafe "unsafe"

	unversioned "github.com/deislabs/ratify/api/unversioned"
	conversion "k8s.io/apimachinery/pkg/conversion"
)

func Convert_unversioned_StoreSpec_To_v1alpha1_StoreSpec(in *unversioned.StoreSpec, out *StoreSpec, _ conversion.Scope) error {
	out.Name = in.Name
	out.Address = in.Address
	out.Source = (*PluginSource)(unsafe.Pointer(in.Source))
	out.Parameters = in.Parameters
	return nil
}

func Convert_unversioned_StoreStatus_To_v1alpha1_StoreStatus(in *unversioned.StoreStatus, out *StoreStatus, _ conversion.Scope) error {
	return nil
}

func Convert_unversioned_VerifierSpec_To_v1alpha1_VerifierSpec(in *unversioned.VerifierSpec, out *VerifierSpec, _ conversion.Scope) error {
	out.Name = in.Name
	out.Address = in.Address
	out.ArtifactTypes = in.ArtifactTypes
	out.Source = (*PluginSource)(unsafe.Pointer(in.Source))
	return nil
}
