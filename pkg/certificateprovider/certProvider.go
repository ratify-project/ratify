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

package certificateprovider

import (
	"context"

	"github.com/deislabs/ratify/pkg/certificateprovider/azurekeyvault/types"
)

type CertificateProvider interface {
	GetCertificatesContent(ctx context.Context, attrib map[string]string) ([]types.CertificateFile, error)
}

// CRD manager call this method to fetch certificate in memory
func SetCertificate(ctx context.Context, certStoreName string, attrib map[string]string) error {
	// Not yet implemented
	return nil
}

// CRD manager call this method to remove certificate from map
func DeleteCertificate(ctx context.Context, certStoreName string) error {
	// Not yet implemented
	return nil
}

// Verifier call this method to get validation certificate
func GetCertificate(ctx context.Context, certStoreName string) ([]types.CertificateFile, error) {
	// TO implement
	return nil, nil
}
