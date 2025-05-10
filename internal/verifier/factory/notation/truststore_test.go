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
	"crypto/x509/pkix"
	"testing"

	"github.com/notaryproject/notation-go/verifier/truststore"
)

const (
	storeName1 = "store1"
	storeName2 = "store2"
)

func TestTrustStore(t *testing.T) {
	trustStore := newTrustStore()

	cert1 := &x509.Certificate{Subject: pkix.Name{CommonName: "cert1"}}
	cert2 := &x509.Certificate{Subject: pkix.Name{CommonName: "cert2"}}

	trustStore.addCertificates(truststore.TypeCA, storeName1, []*x509.Certificate{cert1, cert2})

	certs, err := trustStore.GetCertificates(context.Background(), truststore.TypeCA, storeName1)
	if err != nil {
		t.Fatalf("failed to get certificates: %v", err)
	}
	if len(certs) != 2 {
		t.Fatalf("expected 2 certificates, got %d", len(certs))
	}

	certs, err = trustStore.GetCertificates(context.Background(), truststore.TypeCA, storeName2)
	if err != nil {
		t.Fatalf("failed to get certificates: %v", err)
	}
	if certs != nil {
		t.Fatalf("expected no certificates, got %d", len(certs))
	}
}
