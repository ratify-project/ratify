package certificateprovider

import (
	"context"

	"github.com/deislabs/ratify/pkg/certificateprovider/azurekeyvault/types"
)

type CertificateProvider interface {
	GetCertificatesContent(ctx context.Context, attrib map[string]string) ([]types.CertificateFile, error)
}

// CRD manager call this method to fetch certificate in memory
func SetCertificate(ctx context.Context, certStoreName string, attrib map[string]string) {
	// To implement
}

// CRD manager call this method to remove certificate from map
func DeleteCertificate(ctx context.Context, certStoreName string) {
	// To implement
}

// Verifier call this method to get validation certificate
func GetCertificate(ctx context.Context, certStoreName string) ([]types.CertificateFile, error) {
	// TO implement
	return nil, nil
}
