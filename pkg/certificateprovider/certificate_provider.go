package certificateprovider

import (
	"context"

	"github.com/deislabs/ratify/pkg/certificateprovider/azurekeyvault/types"
)

// This is a map containing Cert store configuration including name, tenantID, and cert object information
type CertStoreConfig map[string]string

// CertificateProvider is an interface that defines methods to be implemented by a each certificate provider
type CertificateProvider interface {
	// Returns an array of certificates based on certificate properties defined in attrib map
	GetCertificatesContent(ctx context.Context, attrib map[string]string) ([]types.Certificate, error)
}
