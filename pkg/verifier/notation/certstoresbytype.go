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
	"fmt"

	"github.com/notaryproject/notation-go/verifier/truststore"
	"github.com/ratify-project/ratify/internal/logger"
)

type certStoreType string

const (
	CA               certStoreType = "CA"
	SigningAuthority certStoreType = "signingAuthority"
)

func (certstoretype certStoreType) String() string {
	return string(certstoretype)
}

// verificationCertStores describes the configuration of verification certStores
// type verificationCertStores supports new format map[string]map[string][]string
//
//	{
//	  "ca": {
//	    "certs": {"kv1", "kv2"},
//	  },
//	  "signingauthority": {
//	    "certs": {"kv3"}
//	  },
//	}
//
// type verificationCertStores supports legacy format map[string][]string as well.
//
//	{
//	  "certs": {"kv1", "kv2"},
//	},
type verificationCertStores map[string]interface{}

// certStoresByType implements certStores interface and place certs under the trustStoreType
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

// newCertStoreByType do type assertion and convert certStores configuration into certStoresByType
func newCertStoreByType(confInNewFormat verificationCertStores) (certStores, error) {
	s := make(certStoresByType)
	for certstoretype, certStores := range confInNewFormat {
		s[certStoreType(certstoretype)] = make(map[string][]string)
		storesMapOfInterface, ok := certStores.(verificationCertStores)
		if !ok {
			return nil, fmt.Errorf("certStores: %s assertion to type verificationCertStores failed", certStores)
		}
		for certStore, certProviders := range storesMapOfInterface {
			var certProvidersStr []string
			convertedCerts, ok := certProviders.([]interface{})
			if !ok {
				return nil, fmt.Errorf("certProviders: %s assertion to type []interface{} failed", certProviders)
			}
			for _, certProvider := range convertedCerts {
				certProviderStr, ok := certProvider.(string)
				if !ok {
					return nil, fmt.Errorf("certProvider: %s assertion to type string failed", certProvider)
				}
				certProvidersStr = append(certProvidersStr, certProviderStr)
			}
			s[certStoreType(certstoretype)][certStore] = certProvidersStr
		}
	}
	return s, nil
}

// GetCertGroup returns certain type of certs from namedStore
func (s certStoresByType) GetCertGroup(ctx context.Context, storeType truststore.Type, namedStore string) (certGroup []string) {
	if certStores, ok := s[certStoreType(storeType)]; ok {
		if certGroup, ok = certStores[namedStore]; ok {
			return
		}
	}
	logger.GetLogger(ctx, logOpt).Warnf("unable to fetch certGroup from namedStore: %+v in type: %v", namedStore, storeType)
	return
}
