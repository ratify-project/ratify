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

	"github.com/deislabs/ratify/internal/logger"
	"github.com/notaryproject/notation-go/verifier/truststore"
)

type verificationCertStores map[string]interface{}

// type certStores map[string][]string
type certStoresByType map[string]map[string][]string

func newCertStoreByType(confVerificationCertStores verificationCertStores) (certStoresByType, error) {
	certStoresByType := make(map[string]map[string][]string)
	for certStoreType, certStores := range confVerificationCertStores {
		certStoresByType[certStoreType] = make(map[string][]string)
		for certStore, certs := range certStores.(verificationCertStores) {
			var reformedCerts []string
			for _, cert := range certs.([]interface{}) {
				if reformedCert, ok := cert.(string); ok {
					reformedCerts = append(reformedCerts, reformedCert)
				}
			}
			certStoresByType[certStoreType][certStore] = reformedCerts
		}
	}
	return certStoresByType, nil
}

// GetCertGroupFromStore returns certain type of certs from namedStore
func GetCertGroupFromStore(ctx context.Context, certStoresByType certStoresByType, storeType truststore.Type, namedStore string) (certGroup []string) {
	if certStores, ok := certStoresByType[string(storeType)]; ok {
		if certGroup, ok = certStores[namedStore]; ok {
			return
		}
	}
	logger.GetLogger(ctx, logOpt).Debugf("unable to fetch certGroup from namedStore: %+v in type: %v", namedStore, storeType)
	return
}
