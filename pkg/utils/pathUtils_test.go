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
	"crypto/x509"
	"os"

	"path/filepath"

	"github.com/notaryproject/notation-go-lib/crypto/cryptoutil"
)

func GetCertificatesFromPath(path string) ([]*x509.Certificate, error) {

	var certs []*x509.Certificate

	err := filepath.Walk(path, func(file string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			cert, certError := cryptoutil.ReadCertificateFile(file) // this method returns empty if file was not a certificate file
			certs = append(certs, cert...)
			if certError != nil {
				return certError
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return certs, nil
}
