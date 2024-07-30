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
	"crypto"
	"crypto/x509"

	"github.com/ratify-project/ratify/pkg/keymanagementprovider"
	"github.com/ratify-project/ratify/pkg/keymanagementprovider/config"
)

type TestKeyManagementProviderFactory struct {
}

func (f *TestKeyManagementProviderFactory) Create(_ string, _ config.KeyManagementProviderConfig, _ string) (keymanagementprovider.KeyManagementProvider, error) {
	var certMap map[keymanagementprovider.KMPMapKey][]*x509.Certificate
	var keyMap map[keymanagementprovider.KMPMapKey]crypto.PublicKey
	return &TestKeyManagementProvider{certificates: certMap, keys: keyMap}, nil
}
