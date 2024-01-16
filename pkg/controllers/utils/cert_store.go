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
	"encoding/json"
	"fmt"

	ctxUtils "github.com/deislabs/ratify/internal/context"
	"github.com/deislabs/ratify/pkg/certificateprovider"
	"github.com/deislabs/ratify/pkg/controllers"
	"github.com/deislabs/ratify/pkg/utils"
	"github.com/sirupsen/logrus"
)

func GetCertStoreConfig(raw []byte) (map[string]string, error) {
	attributes := map[string]string{}

	if string(raw) == "" {
		return nil, fmt.Errorf("received empty parameters")
	}

	if err := json.Unmarshal(raw, &attributes); err != nil {
		logrus.Error(err, ",unable to decode cert store parameters", "Parameters.Raw", raw)
		return attributes, err
	}

	return attributes, nil
}

// given the name of the target provider, returns the provider from the providers map
func GetCertificateProvider(providers map[string]certificateprovider.CertificateProvider, providerName string) (certificateprovider.CertificateProvider, error) {
	providerName = utils.TrimSpaceAndToLower(providerName)
	provider, registered := providers[providerName]
	if !registered {
		return nil, fmt.Errorf("Unknown provider value '%v' defined", provider)
	}
	return provider, nil
}

// returns the internal certificate map
func GetCertificatesMap(ctx context.Context) map[string][]*x509.Certificate {
	return controllers.CertificatesMap.GetCertStores(ctxUtils.GetNamespace(ctx))
}
