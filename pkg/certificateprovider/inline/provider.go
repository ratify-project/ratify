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

package inline

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

const (
	// ValueParameter is the name of the parameter that contains the certificate (chain) as a string in PEM format
	ValueParameter = "value"
)

func GetCertificates(ctx context.Context, attrib map[string]string) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate

	value, ok := attrib[ValueParameter]
	if !ok {
		return nil, fmt.Errorf("value parameter is not set")
	}

	block, rest := pem.Decode([]byte(value))
	if block == nil && len(rest) > 0 {
		return nil, fmt.Errorf("failed to decode pem block")
	}

	for block != nil {
		if block.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("error parsing x509 certificate: %w", err)
			}
			certs = append(certs, cert)
			block, rest = pem.Decode(rest)
			if block == nil && len(rest) > 0 {
				return nil, fmt.Errorf("failed to decode pem block")
			}
		}
	}

	return certs, nil
}
