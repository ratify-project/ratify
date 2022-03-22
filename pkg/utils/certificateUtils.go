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
	"strings"

	"path/filepath"

	"github.com/notaryproject/notation-go-lib/crypto/cryptoutil"
	"github.com/sirupsen/logrus"
)

const (
	homeDirectoryUnixShortcut = "~"
	homeDirectoryUnixPrefix   = homeDirectoryUnixShortcut + "/"
)

func GetCertificatesFromPath(path string) ([]*x509.Certificate, error) {

	var certs []*x509.Certificate
	path = ReplaceHomeShortcut(path)
	err := filepath.Walk(path, func(file string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() {
			cert, certError := cryptoutil.ReadCertificateFile(file) // ReadCertificateFile returns empty if file was not a certificate
			if certError != nil {
				return certError
			}
			if cert != nil {
				certs = append(certs, cert...)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	logrus.Infof("%v notary verification certificates loaded from path '%v'", len(certs), path)
	return certs, nil
}

func ReplaceHomeShortcut(path string) string {
	if strings.HasPrefix(path, homeDirectoryUnixPrefix) {
		home, err := os.UserHomeDir()
		if err == nil && len(home) > 0 {
			return strings.Replace(path, homeDirectoryUnixShortcut, home, 1) // replace 1 instance
		} else {
			logrus.Warningf("Path replacement failed, error %v, value of Home dir %v", err, home)
		}
	}
	return path
}
