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

package utils

import (
	"context"
	"crypto/x509"
	"testing"

	"github.com/deislabs/ratify/pkg/controllers"

	ctxUtils "github.com/deislabs/ratify/internal/context"
	cs "github.com/deislabs/ratify/pkg/customresources/certificatestores"
)

func TestGetCertificatesMap(t *testing.T) {
	controllers.CertificatesMap = cs.NewActiveCertStores()
	controllers.CertificatesMap.AddStore("default", "default/certStore", []*x509.Certificate{})
	ctx := ctxUtils.SetContextWithNamespace(context.Background(), "default")

	if certs := GetCertificatesMap(ctx); len(certs) != 1 {
		t.Fatalf("Expected 1 certificate store, got %d", len(certs))
	}
}
