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

package mocks

import (
	"context"
	"crypto/x509"

	"github.com/deislabs/ratify/pkg/keymanagementprovider"
)

type TestKeyManagementProvider struct {
	certificates map[keymanagementprovider.KMPMapKey][]*x509.Certificate
	status       keymanagementprovider.KeyManagementProviderStatus
	err          error
}

func (c *TestKeyManagementProvider) GetCertificates(_ context.Context) (map[keymanagementprovider.KMPMapKey][]*x509.Certificate, keymanagementprovider.KeyManagementProviderStatus, error) {
	return c.certificates, c.status, c.err
}
