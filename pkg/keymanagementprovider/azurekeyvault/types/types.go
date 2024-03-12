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

const (
	// Static string for certificate type
	CertificateType = "CERTIFICATE"
	// key of the certificate status property
	CertificatesStatus = "Certificates"
	// Static string for certificate name for the certificate status property
	CertificateName = "Name"
	// Certificate version string for the certificate status property
	CertificateVersion = "Version"
	// Last refreshed string for the certificate status property
	CertificateLastRefreshed = "LastRefreshed"
)

// KeyVaultCertificate holds keyvault certificate related config
type KeyVaultCertificate struct {
	// the name of the Azure Key Vault certificate
	Name string `json:"name" yaml:"name"`
	// the version of the Azure Key Vault certificate
	Version string `json:"version" yaml:"version"`
}
