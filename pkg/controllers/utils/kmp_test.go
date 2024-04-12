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

	ctxUtils "github.com/deislabs/ratify/internal/context"
	"github.com/deislabs/ratify/pkg/controllers"
	kmp "github.com/deislabs/ratify/pkg/customresources/keymanagementproviders"
	"github.com/deislabs/ratify/pkg/keymanagementprovider"
)

func TestGetKMPCertificates(t *testing.T) {
	kmpCerts := map[keymanagementprovider.KMPMapKey][]*x509.Certificate{
		{
			Name:    "testName",
			Version: "testVersion",
		}: {},
	}
	controllers.KMPCertificateMap = kmp.NewActiveCertStores()
	controllers.KMPCertificateMap.AddCerts("default", "default/certStore", kmpCerts)
	ctx := ctxUtils.SetContextWithNamespace(context.Background(), "default")

	if certs := GetKMPCertificates(ctx, "default/certStore"); len(certs) != 0 {
		t.Fatalf("Expected 0 certificate, got %d", len(certs))
	}
}
