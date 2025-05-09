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

package notation

import (
	"context"
	"crypto/x509"

	"github.com/notaryproject/notation-go/verifier/truststore"
	"github.com/sirupsen/logrus"
)

type trustStore struct {
	stores map[truststore.Type]map[string][]*x509.Certificate
}

func newTrustStore() *trustStore {
	return &trustStore{
		stores: make(map[truststore.Type]map[string][]*x509.Certificate),
	}
}

// GetCertificates implements [truststore.X509TrustStore] interface.
func (s *trustStore) GetCertificates(_ context.Context, storeType truststore.Type, namedStore string) ([]*x509.Certificate, error) {
	logrus.Infof("Getting certificates from trust store %s", namedStore)
	if namedStores, ok := s.stores[storeType]; ok {
		if certs, ok := namedStores[namedStore]; ok {
			logrus.Infof("Found %d certificates in trust store %s", len(certs), namedStore)
			return certs, nil
		}
	}
	return nil, nil
}

// addCertificates adds provided certificates to the trust store.
func (s *trustStore) addCertificates(storeType truststore.Type, namedStore string, certs []*x509.Certificate) {
	if s.stores[storeType] == nil {
		s.stores[storeType] = make(map[string][]*x509.Certificate)
	}
	s.stores[storeType][namedStore] = append(s.stores[storeType][namedStore], certs...)
}
