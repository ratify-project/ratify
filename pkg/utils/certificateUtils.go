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
	"io/fs"
	"os"
	"strings"

	"path/filepath"

	"github.com/deislabs/ratify/pkg/homedir"
	"github.com/notaryproject/notation-go-lib/crypto/cryptoutil"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func GetCertificatesFromPath(path string) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate
	fileMap := map[string]bool{} //a map to track path of physical files

	path = ReplaceHomeShortcut(path)

	err := filepath.Walk(path, func(file string, info os.FileInfo, err error) error {

		physicalFileInfo := info
		physicalFilePath := file

		if info == nil || err != nil {
			logrus.Warnf("Invalid path '%v' skipped, error %v", file, err)
			return nil
		}

		// In a cluster environment, each mounted file results in a physical file () and a symlink ()
		// check if file is a link and get the actual file path
		if info.Mode()&os.ModeSymlink != 0 {

			physicalFilePath, err = filepath.EvalSymlinks(file)

			if err != nil {
				logrus.Errorf("error evaluating symbolic link %v , error '%v'", file, file)
				return nil
			}

			physicalFileInfo, err = os.Lstat(physicalFilePath)

			if err != nil {
				logrus.Errorf("error getting file info for path '%v', error '%v'", physicalFilePath, err)
				return nil
			}
		}

		if !physicalFileInfo.IsDir() {
			if _, ok := fileMap[physicalFilePath]; !ok {
				err = loadCertFile(physicalFileInfo, physicalFilePath, &certs, fileMap)
				if err != nil {
					return errors.Wrap(err, "error reading certificate file "+physicalFilePath)
				}
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

func loadCertFile(fileInfo fs.FileInfo, filePath string, certificate *[]*x509.Certificate, fileMap map[string]bool) error {

	cert, certError := cryptoutil.ReadCertificateFile(filePath) // ReadCertificateFile returns empty if file was not a certificate
	if certError != nil {
		return certError
	}
	if cert != nil {
		*certificate = append(*certificate, cert...)
		fileMap[filePath] = true
	}

	return nil
}

func ReplaceHomeShortcut(path string) string {

	shortcutPrefix := homedir.GetShortcutString() + string(os.PathSeparator)
	if strings.HasPrefix(path, shortcutPrefix) {
		home := homedir.Get()
		if len(home) > 0 {
			return strings.Replace(path, homedir.GetShortcutString(), home, 1) // replace 1 instance
		} else {
			logrus.Warningf("Path '%v' replacement failed , value of Home dir '%v'", path, home)
		}
	}
	return path
}
