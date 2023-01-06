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
