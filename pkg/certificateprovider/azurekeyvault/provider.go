package azurekeyvault

// This class is based on implementation from  azure secret store csi provider
// Source: https://github.com/Azure/secrets-store-csi-driver-provider-azure/blob/master/pkg/provider/
import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	kv "github.com/Azure/azure-sdk-for-go/services/keyvault/v7.1/keyvault"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/deislabs/ratify/pkg/certificateprovider/azurekeyvault/types"
	"github.com/sirupsen/logrus"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type keyvaultObject struct {
	content string
	version string
}

func GetCertificatesContent(ctx context.Context, attrib map[string]string) ([]types.CertificateFile, error) {

	keyvaultName := types.GetKeyVaultName(attrib)
	cloudName := types.GetCloudName(attrib)
	tenantID := types.GetTenantID(attrib)
	workloadIdentityClientID := types.GetClientID(attrib)

	if keyvaultName == "" {
		return nil, fmt.Errorf("keyvaultName is not set")
	}
	if tenantID == "" {
		return nil, fmt.Errorf("tenantID is not set")
	}
	if workloadIdentityClientID == "" {
		return nil, fmt.Errorf("clientId is not set")
	}

	azureCloudEnv, err := ParseAzureEnvironment(cloudName)
	if err != nil {
		return nil, fmt.Errorf("cloudName %s is not valid, error: %w", cloudName, err)
	}

	// 1. cleaning up keyvault objects definition
	objectsStrings := types.GetCertificates(attrib)
	if objectsStrings == "" {
		return nil, fmt.Errorf("objects is not set")
	}
	logrus.Infof("objects string defined in secret provider class, objects %v", objectsStrings)

	objects, err := types.GetCertificatesArray(objectsStrings)
	if err != nil {
		return nil, fmt.Errorf("failed to yaml unmarshal objects, error: %w", err)
	}
	logrus.Infof("unmarshaled objects yaml, objectsArray %v", objects.Array)

	keyVaultCerts := []types.KeyVaultCertificate{}
	for i, object := range objects.Array {
		var keyVaultCert types.KeyVaultCertificate
		err = yaml.Unmarshal([]byte(object), &keyVaultCert)
		if err != nil {
			return nil, fmt.Errorf("unmarshal failed for keyVaultCerts at index %d, error: %w", i, err)
		}
		// remove whitespace from all fields in keyVaultCert
		formatKeyVaultCertificate(&keyVaultCert)
		if err = validate(keyVaultCert); err != nil {
			return nil, wrapObjectTypeError(err, keyVaultCert.CertificateName, keyVaultCert.CertificateVersion)
		}

		keyVaultCerts = append(keyVaultCerts, keyVaultCert)
	}

	logrus.Infof("unmarshaled key vault objects, keyVaultObjects %v , count %v", keyVaultCerts, len(keyVaultCerts))

	if len(keyVaultCerts) == 0 {
		return nil, nil
	}

	// 2. initialize keyvault client

	vaultURL, err := getVaultURL(keyvaultName, azureCloudEnv.KeyVaultDNSSuffix)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get vault")
	}
	logrus.Infof("vaultName %v, vaultURL %v", keyvaultName, *vaultURL)

	kvClient, err := initializeKvClient(ctx, azureCloudEnv.KeyVaultEndpoint, azureCloudEnv.ActiveDirectoryEndpoint, tenantID, workloadIdentityClientID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get keyvault client")
	}

	// 3. for each object , get content bytes
	files := []types.CertificateFile{}
	for _, keyVaultCert := range keyVaultCerts {
		logrus.Infof("fetching object from key vault, certName %v,  keyvault %v", keyVaultCert.CertificateName, keyvaultName)

		resolvedKvCerts, err := resolveCertificateVersions(ctx, kvClient, keyVaultCert, *vaultURL)
		if err != nil {
			return nil, err
		}

		for _, resolvedKvCert := range resolvedKvCerts {
			// fetch the object from Key Vault
			result, err := getCertificate(ctx, kvClient, *vaultURL, resolvedKvCert)
			if err != nil {
				return nil, err
			}

			for idx := range result {
				r := result[idx]
				objectContent := []byte(r.content)

				file := types.CertificateFile{
					Path:    resolvedKvCert.GetFileName(),
					Content: objectContent,
					Version: r.version,
				}

				files = append(files, file)
				logrus.Infof("added file %v to response file", file.Path)
			}
		}
	}
	return files, nil
}

// formatKeyVaultCertificate formats the fields in KeyVaultCertificate
func formatKeyVaultCertificate(object *types.KeyVaultCertificate) {
	if object == nil {
		return
	}
	objectPtr := reflect.ValueOf(object)
	objectValue := objectPtr.Elem()

	for i := 0; i < objectValue.NumField(); i++ {
		field := objectValue.Field(i)
		if field.Type() != reflect.TypeOf("") {
			continue
		}
		str := field.Interface().(string)
		str = strings.TrimSpace(str)
		field.SetString(str)
	}
}

// ParseAzureEnvironment returns azure environment by name
func ParseAzureEnvironment(cloudName string) (*azure.Environment, error) {
	var env azure.Environment
	var err error
	if cloudName == "" {
		env = azure.PublicCloud
	} else {
		env, err = azure.EnvironmentFromName(cloudName)
	}
	return &env, err
}

func getVaultURL(keyvaultName string, KeyVaultDNSSuffix string) (vaultURL *string, err error) {
	// Key Vault name must be a 3-24 character string
	if len(keyvaultName) < 3 || len(keyvaultName) > 24 {
		return nil, errors.Errorf("Invalid vault name: %q, must be between 3 and 24 chars", keyvaultName)
	}
	// See docs for validation spec: https://docs.microsoft.com/en-us/azure/key-vault/about-keys-secrets-and-certificates#objects-identifiers-and-versioning
	isValid := regexp.MustCompile(`^[-A-Za-z0-9]+$`).MatchString
	if !isValid(keyvaultName) {
		return nil, errors.Errorf("Invalid vault name: %q, must match [-a-zA-Z0-9]{3,24}", keyvaultName)
	}

	vaultDNSSuffixValue := KeyVaultDNSSuffix
	vaultURI := "https://" + keyvaultName + "." + vaultDNSSuffixValue + "/"
	return &vaultURI, nil
}

func initializeKvClient(ctx context.Context, KeyVaultEndpoint string, aadEndpoint string, tenantID string, clientId string) (*kv.BaseClient, error) {
	kvClient := kv.New()
	kvEndpoint := strings.TrimSuffix(KeyVaultEndpoint, "/")

	err := kvClient.AddToUserAgent("ratify")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to add user agent to keyvault client")
	}

	kvClient.Authorizer, err = getAuthorizerForWorkloadIdentity(ctx, tenantID, clientId, kvEndpoint, aadEndpoint)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to get authorizer for keyvault client")
	}
	return &kvClient, nil
}

/*
Given a base key vault object and a list of object versions and their created dates, find
the latest kvObject.ObjectVersionHistory versions and return key vault objects with the
appropriate alias and version.
The alias is determine by the index of the version starting with 0 at the specified version (or
latest if no version is specified).
*/
func getLatestNKeyVaultObjects(kvCert types.KeyVaultCertificate, kvObjectVersions types.KeyVaultObjectVersionList) []types.KeyVaultCertificate {
	baseFileName := kvCert.GetFileName()
	objects := []types.KeyVaultCertificate{}

	sort.Sort(kvObjectVersions)

	// if we're being asked for the latest, then there's no need to skip any versions
	foundFirst := kvCert.CertificateVersion == "" || kvCert.CertificateVersion == "latest"

	for _, objectVersion := range kvObjectVersions {
		foundFirst = foundFirst || objectVersion.Version == kvCert.CertificateVersion

		if foundFirst {
			length := len(objects)
			newObject := kvCert

			newObject.CertificateAlias = filepath.Join(baseFileName, strconv.Itoa(length))
			newObject.CertificateVersion = objectVersion.Version

			objects = append(objects, newObject)

			if length+1 > int(kvCert.CertificateVersionHistory) {
				break
			}
		}
	}

	return objects
}

func getKeyVaultCertificateVersions(ctx context.Context, kvClient *kv.BaseClient, kvObject types.KeyVaultCertificate, vaultURL string) (versions types.KeyVaultObjectVersionList, err error) {

	return getCertificateVersions(ctx, kvClient, vaultURL, kvObject)

}

func resolveCertificateVersions(ctx context.Context, kvClient *kv.BaseClient, kvObject types.KeyVaultCertificate, vaultURL string) (versions []types.KeyVaultCertificate, err error) {
	if kvObject.IsSyncingSingleVersion() {
		// version history less than or equal to 1 means only sync the latest and
		// don't add anything to the file name
		return []types.KeyVaultCertificate{kvObject}, nil
	}

	kvObjectVersions, err := getCertificateVersions(ctx, kvClient, vaultURL, kvObject)
	if err != nil {
		return nil, err
	}

	return getLatestNKeyVaultObjects(kvObject, kvObjectVersions), nil
}

func getCertificateVersions(ctx context.Context, kvClient *kv.BaseClient, vaultURL string, kvObject types.KeyVaultCertificate) ([]types.KeyVaultObjectVersion, error) {
	kvVersionsList, err := kvClient.GetCertificateVersions(ctx, vaultURL, kvObject.CertificateName, nil)
	if err != nil {
		return nil, wrapObjectTypeError(err, kvObject.CertificateName, kvObject.CertificateVersion)
	}

	certVersions := types.KeyVaultObjectVersionList{}

	for notDone := true; notDone; notDone = kvVersionsList.NotDone() {
		for _, cert := range kvVersionsList.Values() {
			if cert.Attributes != nil {
				objectVersion := getObjectVersion(*cert.ID)
				created := date.UnixEpoch()

				if cert.Attributes.Created != nil {
					created = time.Time(*cert.Attributes.Created)
				}

				if cert.Attributes.Enabled != nil && *cert.Attributes.Enabled {
					certVersions = append(certVersions, types.KeyVaultObjectVersion{
						Version: objectVersion,
						Created: created,
					})
				}
			}
		}

		err = kvVersionsList.NextWithContext(ctx)
		if err != nil {
			return nil, wrapObjectTypeError(err, kvObject.CertificateName, kvObject.CertificateVersion)
		}
	}

	return certVersions, nil
}

// getCertificate retrieves the certificate from the vault
func getCertificate(ctx context.Context, kvClient *kv.BaseClient, vaultURL string, kvObject types.KeyVaultCertificate) ([]keyvaultObject, error) {
	// for object type "cert" the certificate is written to the file in PEM format
	certbundle, err := kvClient.GetCertificate(ctx, vaultURL, kvObject.CertificateName, kvObject.CertificateVersion)
	if err != nil {
		return nil, wrapObjectTypeError(err, kvObject.CertificateName, kvObject.CertificateVersion)
	}
	if certbundle.Cer == nil {
		return nil, errors.Errorf("certificate value is nil")
	}
	if certbundle.ID == nil {
		return nil, errors.Errorf("certificate id is nil")
	}
	version := getObjectVersion(*certbundle.ID)

	certBlock := &pem.Block{
		Type:  types.CertificateType,
		Bytes: *certbundle.Cer,
	}
	var pemData []byte
	pemData = append(pemData, pem.EncodeToMemory(certBlock)...)
	return []keyvaultObject{{content: string(pemData), version: version}}, nil
}

// ParsePEMCertificates parses PEM/DER encoded certificates from
// the given PEM data.
func ParsePEMCertificates(pemData []byte) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate
	for {
		var der *pem.Block
		der, pemData = pem.Decode(pemData)
		if der == nil {
			break
		}
		if der.Type == "CERTIFICATE" {
			dcerts, err := x509.ParseCertificates(der.Bytes)
			if err != nil {
				return nil, err
			}
			certs = append(certs, dcerts...)
		}
	}
	return certs, nil
}

func wrapObjectTypeError(err error, objectName, objectVersion string) error {
	return errors.Wrapf(err, "failed to get certificate objectName:%s, objectVersion:%s", objectName, objectVersion)
}

// getObjectVersion parses the id to retrieve the version
// of object fetched
// example id format - https://kindkv.vault.azure.net/secrets/actual/1f304204f3624873aab40231241243eb
// TODO (aramase) follow up on https://github.com/Azure/azure-rest-api-specs/issues/10825 to provide
// a native way to obtain the version
func getObjectVersion(id string) string {
	splitID := strings.Split(id, "/")
	return splitID[len(splitID)-1]
}
