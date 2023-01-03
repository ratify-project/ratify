package certificateprovider

import (
	"context"

	"github.com/deislabs/ratify/pkg/certificateprovider/akv/types"
)

var (
	// a map to track active stores
	CertList = map[string]string{}
)

func GetCert(ctx context.Context) ([]types.SecretFile, error) {

	//(ctx context.Context, attrib map[string]string, defaultFilePermission os.FileMode)
	// TODO: populate the map with keyvault info
	attrib := map[string]string{}
	attrib["keyvaultName"] = "notarycerts"
	attrib["clientID"] = "1c7ac023-5bf6-4916-83f2-96dd203e35a3"
	attrib["cloudName"] = "AzurePublicCloud"
	attrib["tenantID"] = "72f988bf-86f1-41af-91ab-2d7cd011db47"

	attrib["objects"] = "array:\n- |\n  objectName: wabbit-networks-io	\n  objectAlias: \"\"\n  ObjectVersion: 97a1545d893344079ce57699c8810590 \n  objectVersionHistory: 0\n  objectType: cert\n  objectFormat: \"\"\n  objectEncoding: \"\"\n  filePermission: \"\"\n"

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
