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

package keymanagementproviders

import (
	"crypto/x509"
	"testing"

	"github.com/deislabs/ratify/pkg/keymanagementprovider"
	"github.com/deislabs/ratify/pkg/utils"
)

const (
	namespace1 = "namespace1"
	namespace2 = "namespace2"
	name1      = "name1"
	name2      = "name2"
)

func TestCertStoresOperations(t *testing.T) {
	activeCertStores := NewActiveCertStores()

	certStore1 := map[keymanagementprovider.KMPMapKey][]*x509.Certificate{
		{Name: "testName1", Version: "testVersion1"}: {utils.CreateTestCert()},
	}
	certStore2 := map[keymanagementprovider.KMPMapKey][]*x509.Certificate{
		{Name: "testName2", Version: "testVersion2"}: {utils.CreateTestCert()},
	}

	if len(activeCertStores.GetCertStores(namespace1, name1)) != 0 {
		t.Errorf("Expected activeCertStores to have 0 cert store, but got %d", len(activeCertStores.GetCertStores(namespace1, name1)))
	}

	activeCertStores.AddCerts(namespace1, name1, certStore1)
	activeCertStores.AddCerts(namespace2, name2, certStore2)

	if len(activeCertStores.GetCertStores(namespace1, name1)) != 1 {
		t.Errorf("Expected activeCertStores to have 1 cert store, but got %d", len(activeCertStores.GetCertStores(namespace1, name1)))
	}

	if len(activeCertStores.GetCertStores(namespace2, name2)) != 1 {
		t.Errorf("Expected activeCertStores to have 1 cert store, but got %d", len(activeCertStores.GetCertStores(namespace2, name2)))
	}

	activeCertStores.DeleteCerts(namespace1, name1)
	activeCertStores.DeleteCerts(namespace2, name2)

	if len(activeCertStores.GetCertStores(namespace1, name1)) != 0 {
		t.Errorf("Expected activeCertStores to have 0 cert store, but got %d", len(activeCertStores.GetCertStores(namespace1, name1)))
	}

	if len(activeCertStores.GetCertStores(namespace2, name2)) != 0 {
		t.Errorf("Expected activeCertStores to have 0 cert store, but got %d", len(activeCertStores.GetCertStores(namespace2, name2)))
	}
}
