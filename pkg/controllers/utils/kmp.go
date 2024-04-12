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

	ctxUtils "github.com/deislabs/ratify/internal/context"
	"github.com/deislabs/ratify/pkg/controllers"
)

// GetKMPCertificates returns internal certificate map from KMP.
// TODO: returns certificates from both cluster-wide and given namespace as namespaced verifier could access both.
func GetKMPCertificates(ctx context.Context, certStore string) []*x509.Certificate {
	return controllers.KMPCertificateMap.GetCertStores(ctxUtils.GetNamespace(ctx), certStore)
}
