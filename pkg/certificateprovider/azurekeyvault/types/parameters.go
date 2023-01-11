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

package types

// This class is based on implementation from azure secret store csi provider
// Source: https://github.com/Azure/secrets-store-csi-driver-provider-azure/blob/master/pkg/provider/
import (
	"strings"

	"gopkg.in/yaml.v3"
	"k8s.io/klog/v2"
)

// GetKeyVaultName returns the key vault name
func GetKeyVaultName(parameters map[string]string) string {
	return strings.TrimSpace(parameters[KeyVaultNameParameter])
}

// GetCloudName returns the cloud name
func GetCloudName(parameters map[string]string) string {
	return strings.TrimSpace(parameters[CloudNameParameter])
}

// GetTenantID returns the tenant ID
func GetTenantID(parameters map[string]string) string {
	// ref: https://github.com/Azure/secrets-store-csi-driver-provider-azure/issues/857
	tenantID := strings.TrimSpace(parameters["tenantID"])
	if tenantID != "" {
		return tenantID
	}
	klog.V(3).Info("tenantId is deprecated and will be removed in a future release. Use 'tenantID' instead")
	return strings.TrimSpace(parameters[TenantIDParameter])
}

// GetClientID returns the client ID
func GetClientID(parameters map[string]string) string {
	return strings.TrimSpace(parameters[ClientIDParameter])
}

// GetCertificates returns the key vault objects
func GetCertificates(parameters map[string]string) string {
	return strings.TrimSpace(parameters[CertificatesParameter])
}

// GetCertificatesArray returns the key vault objects array
func GetCertificatesArray(objects string) (StringArray, error) {
	var a StringArray
	//var b string // Susan to Fix this
	err := yaml.Unmarshal([]byte(objects), &a)

	return a, err
}

// IsSyncingSingleVersion returns true if the object is configured
// to only sync a single specific version of the secret
func (kv KeyVaultCertificate) IsSyncingSingleVersion() bool {
	return kv.CertificateVersionHistory <= 1
}

// GetFileName returns the file name for the secret
// 1. If the object alias is specified, it will be used
// 2. If the object alias is not specified, the object name will be used
func (kv KeyVaultCertificate) GetFileName() string {
	if kv.CertificateAlias != "" {
		return kv.CertificateAlias
	}
	return kv.CertificateName
}
