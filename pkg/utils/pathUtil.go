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
	"os"
	"strings"

	"path/filepath"
)

func GetCertificatesFromPath(path string) ([]string, error) {

	var files []string

	err := filepath.Walk(path, func(file string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			fileExtension := filepath.Ext(file)
			if strings.EqualFold(fileExtension, ".Crt") { //what other files can be read?
				files = append(files, file)
			}
		}
		return nil // or return err?
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}
