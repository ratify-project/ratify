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
	v1beta1 "github.com/deislabs/ratify/api/v1beta1"
	apiconversion "k8s.io/apimachinery/pkg/conversion"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertFrom converts from the Hub version(v1beta1) to this version.
// nolint:revive
func (dst *CertificateStore) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta1.CertificateStore)

	return Convert_v1beta1_CertificateStore_To_v1alpha1_CertificateStore(src, dst, nil)
}

// ConvertTo converts this version to the Hub version(v1beta1).
// nolint:revive
func (src *CertificateStore) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta1.CertificateStore)

	return Convert_v1alpha1_CertificateStore_To_v1beta1_CertificateStore(src, dst, nil)
}

// Overwrite the generated conversion function.
func Convert_v1beta1_CertificateStoreStatus_To_v1alpha1_CertificateStoreStatus(in *v1beta1.CertificateStoreStatus, out *CertificateStoreStatus, s apiconversion.Scope) error {
	return nil
}


// Overwrite the generated conversion function.
// nolint:revive
func Convert_v1alpha1_CertificateStoreStatus_To_v1beta1_CertificateStoreStatus(in *CertificateStoreStatus, out *v1beta1.CertificateStoreStatus, s apiconversion.Scope) error {
    out.Error = "warning: converted from v1alpha1"
	return autoConvert_v1alpha1_CertificateStoreStatus_To_v1beta1_CertificateStoreStatus(in, out, s)
}