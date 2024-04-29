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
	"context"
	"crypto/x509"
	"os"
	"strings"
	"sync"

	"github.com/deislabs/ratify/internal/constants"
	ctxUtils "github.com/deislabs/ratify/internal/context"
	"github.com/deislabs/ratify/pkg/utils"
	vu "github.com/deislabs/ratify/pkg/verifier/utils"
)

const defaultNamespace = "default"

// ActiveCertStores implements the CertStoreManager interface
type ActiveCertStores struct {
	// scopedCertStores is mapping from cert store name to certificate list.
	// The certificate store name is prefixed with the namespace.
	// Example:
	// {
	//   "namespace1/store1": []*x509.Certificate,
	//   "namespace2/store2": []*x509.Certificate
	// }
	// Note: The namespace "default" is reserved for cluster-wide scenario.
	scopedCertStores sync.Map
}

func NewActiveCertStores() CertStoreManager {
	return &ActiveCertStores{}
}

// GetCertStores fulfills the CertStoreManager interface.
// It returns a list of certificates in the given store.
func (c *ActiveCertStores) GetCertsFromStore(ctx context.Context, storeName string) []*x509.Certificate {
	storeName = prependNamespaceToStoreName(storeName)
	if !isCompatibleNamespace(ctx, storeName) {
		return []*x509.Certificate{}
	}
	if certs, ok := c.scopedCertStores.Load(storeName); ok {
		return certs.([]*x509.Certificate)
	}
	return []*x509.Certificate{}
}

// AddStore fulfills the CertStoreManager interface.
// It adds the given certificate under cert store.
func (c *ActiveCertStores) AddStore(storeName string, cert []*x509.Certificate) {
	c.scopedCertStores.Store(storeName, cert)
}

// DeleteStore fulfills the CertStoreManager interface.
// It deletes the given cert store.
func (c *ActiveCertStores) DeleteStore(storeName string) {
	c.scopedCertStores.Delete(storeName)
}

// Namespaced verifiers could access certStores in the same namespace or "default" namespace.
// Cluster-wide verifier could access all certStores.
// Note: the cluster-wide behavior is different from KMP as we need to keep the behavior backward compatible.
func isCompatibleNamespace(ctx context.Context, storeName string) bool {
	namespace := ctxUtils.GetNamespace(ctx)
	if namespace == constants.EmptyNamespace {
		return true
	}
	return strings.HasPrefix(storeName, namespace+constants.NamespaceSeperator) || strings.HasPrefix(storeName, defaultNamespace+constants.NamespaceSeperator)
}

// prependNamespaceToStoreName prepends namespace to store name if not already present.
// If the namespace where Ratify deployed is not set, `default` namespace will be prepended. However, this case should never happen.
func prependNamespaceToStoreName(storeName string) string {
	if vu.IsNamespacedNamed(storeName) {
		return storeName
	}
	defaultNS := defaultNamespace
	if ns, found := os.LookupEnv(utils.RatifyNamespaceEnvVar); found {
		defaultNS = ns
	}
	return defaultNS + constants.NamespaceSeperator + storeName
}
