package certProvider

import (
	"context"
)

type CertProvider interface {

	// Fetch returns AuthConfig for registry.
	FetchCert(ctx context.Context, artifact string) (string, error)
}

var (
	// a map to track active stores
	CertList = map[string]string{}
)

func GetAkvCertProvider(ctx context.Context) (CertProvider, error) {

	//(ctx context.Context, attrib map[string]string, defaultFilePermission os.FileMode)
	//attrib := map[string]string{}
	//attrib["abc"] = "bcd"

	// probably initialize and save in
	//akv.GetSecretsStoreObjectContent(ctx, attrib)
	return nil, nil
}

/*
// returns the list of certificates from the cert cache for the given certStore
// this doesn't not call the keyvault
func () GetCertForCertStore(string certStoreName) {

}

func () initializeKvClient(ctx context.Context) {
}

// CRD tells provider how to fetch the certs, and this method fetch and stores the certs in a map
//  periodically as well, maybe a loop if versions are not pinned ( note crd may also change)
func () GetCerts(keyvaultname string, certStorename string, authProvider string) {*/
