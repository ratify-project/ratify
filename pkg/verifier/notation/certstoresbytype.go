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

type certStoreType string

const (
	CA               certStoreType = "CA"
	SigningAuthority certStoreType = "signingAuthority"
)

func (certstoretype certStoreType) String() string {
	return string(certstoretype)
}

// verificationCertStores describes the configuration of verificationCertStores
type verificationCertStores map[string]interface{}

// certStoresByType describes the configuration of verificationCertStores by certStoreType
//
//	{
//	  "ca": {
//	    "certs": {"kv1", "kv2"},
//	  },
//	  "signingauthority": {
//	    "certs": {"kv3"}
//	  },
//	}
type certStoresByType map[certStoreType]map[string][]string

func newCertStoreByType(confVerificationCertStores verificationCertStores) (certStoresByType, error) {
	certStoresByType := make(map[certStoreType]map[string][]string)
	for certstoretype, certStores := range confVerificationCertStores {
		certStoresByType[certStoreType(certstoretype)] = make(map[string][]string)
		if convertedCertStores, ok := certStores.(verificationCertStores); ok {
			for certStore, certs := range convertedCertStores {
				var reformedCerts []string
				if convertedCerts, ok := certs.([]interface{}); ok {
					for _, cert := range convertedCerts {
						if reformedCert, ok := cert.(string); ok {
							reformedCerts = append(reformedCerts, reformedCert)
						}
					}
					certStoresByType[certStoreType(certstoretype)][certStore] = reformedCerts
				}
			}
		}
	}
	return certStoresByType, nil
}

// GetCertGroupFromStore returns certain type of certs from namedStore
func GetCertGroupFromStore(ctx context.Context, certStoresByType certStoresByType, storeType truststore.Type, namedStore string) (certGroup []string) {
	if certStores, ok := certStoresByType[certStoreType(storeType)]; ok {
		if certGroup, ok = certStores[namedStore]; ok {
			return
		}
	}
	logger.GetLogger(ctx, logOpt).Debugf("unable to fetch certGroup from namedStore: %+v in type: %v", namedStore, storeType)
	return
}
