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

package azurekeyvault

// This class is based on implementation from  azure secret store csi provider
// Source: https://github.com/Azure/secrets-store-csi-driver-provider-azure/blob/master/pkg/provider/
import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/deislabs/ratify/pkg/certificateprovider/azurekeyvault/types"
)

// validate is a helper function to validate the given object
func validate(kv types.KeyVaultCertificate) error {

	if err := validateFileName(kv.GetFileName()); err != nil {
		return err
	}
	return nil
}

// This validate will make sure fileName:
// 1. is not abs path
// 2. does not contain any '..' elements
// 3. does not start with '..'
// These checks have been implemented based on -
// [validateLocalDescendingPath] https://github.com/kubernetes/kubernetes/blob/master/pkg/apis/core/validation/validation.go#L1158-L1170
// [validatePathNoBacksteps] https://github.com/kubernetes/kubernetes/blob/master/pkg/apis/core/validation/validation.go#L1172-L1186
func validateFileName(fileName string) error {
	if len(fileName) == 0 {
		return fmt.Errorf("file name must not be empty")
	}
	// is not abs path
	if filepath.IsAbs(fileName) {
		return fmt.Errorf("file name must be a relative path")
	}
	// does not have any element which is ".."
	parts := strings.Split(filepath.ToSlash(fileName), "/")
	for _, item := range parts {
		if item == ".." {
			return fmt.Errorf("file name must not contain '..'")
		}
	}
	// fallback logic if .. is missed in the previous check
	if strings.Contains(fileName, "..") {
		return fmt.Errorf("file name must not contain '..'")
	}
	return nil
}
