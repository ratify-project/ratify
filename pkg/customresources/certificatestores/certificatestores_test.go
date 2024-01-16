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

package certificatestores

import (
	"crypto/x509"
	"testing"
)

const (
	namespace1 = "namespace1"
	namespace2 = "namespace2"
	name1      = "name1"
	name2      = "name2"
)

func TestCertStoresOperations(t *testing.T) {
	activeCertStores := NewActiveCertStores()

	if !activeCertStores.IsEmpty() {
		t.Errorf("Expected activeCertStores to be empty")
	}

	certStore1 := []*x509.Certificate{}
	certStore2 := []*x509.Certificate{}

	activeCertStores.AddStore(namespace1, name1, certStore1)
	activeCertStores.AddStore(namespace2, name2, certStore2)

	if activeCertStores.IsEmpty() {
		t.Errorf("Expected activeCertStores to not be empty")
	}

	if len(activeCertStores.GetCertStores(namespace1)) != 1 {
		t.Errorf("Expected activeCertStores to have 1 cert store")
	}

	activeCertStores.DeleteStore(namespace1, name1)
	activeCertStores.DeleteStore(namespace2, name2)

	if !activeCertStores.IsEmpty() {
		t.Errorf("Expected activeCertStores to be empty")
	}
}
