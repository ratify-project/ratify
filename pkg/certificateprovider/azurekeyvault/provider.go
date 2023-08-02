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
// Source: https://github.com/Azure/secrets-store-csi-driver-provider-azure/tree/release-1.4/pkg/provider
import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/deislabs/ratify/pkg/certificateprovider"
	"github.com/deislabs/ratify/pkg/certificateprovider/azurekeyvault/types"
	"github.com/deislabs/ratify/pkg/metrics"

	kv "github.com/Azure/azure-sdk-for-go/services/keyvault/v7.1/keyvault"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	providerName      string = "azurekeyvault"
	PKCS12ContentType string = "application/x-pkcs12"
	PEMContentType    string = "application/x-pem-file"
)

type akvCertProvider struct{}

// init calls to register the provider
func init() {
	certificateprovider.Register(providerName, Create())
}

func Create() certificateprovider.CertificateProvider {
	// returning a simple provider for now, overtime we will add metrics and other related properties
	return &akvCertProvider{}
}

// returns an array of certificates based on certificate properties defined in attrib map
// get certificate retrieve the entire cert chain using getSecret API call
func (s *akvCertProvider) GetCertificates(ctx context.Context, attrib map[string]string) ([]*x509.Certificate, certificateprovider.CertificatesStatus, error) {
	keyvaultURI := types.GetKeyVaultURI(attrib)
	cloudName := types.GetCloudName(attrib)
	tenantID := types.GetTenantID(attrib)
	workloadIdentityClientID := types.GetClientID(attrib)

	if keyvaultURI == "" {
		return nil, nil, fmt.Errorf("keyvaultUri is not set")
	}
	if tenantID == "" {
		return nil, nil, fmt.Errorf("tenantID is not set")
	}
	if workloadIdentityClientID == "" {
		return nil, nil, fmt.Errorf("clientID is not set")
	}

	azureCloudEnv, err := parseAzureEnvironment(cloudName)
	if err != nil {
		return nil, nil, fmt.Errorf("cloudName %s is not valid, error: %w", cloudName, err)
	}

	keyVaultCerts, err := getKeyvaultRequestObj(attrib)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get keyvault request object from provider attributes, error: %w", err)
	}

	if len(keyVaultCerts) == 0 {
		return nil, nil, fmt.Errorf("no keyvault certificate configured")
	}

	logrus.Debugf("vaultURI %s", keyvaultURI)

	kvClient, err := initializeKvClient(ctx, azureCloudEnv.KeyVaultEndpoint, tenantID, workloadIdentityClientID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get keyvault client, error: %w", err)
	}

	certs := []*x509.Certificate{}
	certsStatus := []map[string]string{}
	for _, keyVaultCert := range keyVaultCerts {
		logrus.Debugf("fetching secret from key vault, certName %v,  keyvault %v", keyVaultCert.CertificateName, keyvaultURI)

		// fetch the object from Key Vault
		startTime := time.Now()
		secretBundle, err := kvClient.GetSecret(ctx, keyvaultURI, keyVaultCert.CertificateName, keyVaultCert.CertificateVersion)

		if err != nil {
			return nil, nil, fmt.Errorf("failed to get secret objectName:%s, objectVersion:%s, error: %w", keyVaultCert.CertificateName, keyVaultCert.CertificateVersion, err)
		}

		certResult, certProperty, err := getCertsFromSecretBundle(secretBundle, keyVaultCert.CertificateName)

		if err != nil {
			return nil, nil, fmt.Errorf("failed to get certificates from secret bundle:%w", err)
		}

		metrics.ReportAKVCertificateDuration(ctx, time.Since(startTime).Milliseconds(), keyVaultCert.CertificateName)
		certs = append(certs, certResult...)
		certsStatus = append(certsStatus, certProperty...)
	}

	return certs, getCertStatusMap(certsStatus), nil
}

// azure keyvault provider certificate status is a map from "certificates" key to an array of of certificate status
func getCertStatusMap(certsStatus []map[string]string) certificateprovider.CertificatesStatus {
	status := certificateprovider.CertificatesStatus{}
	status[types.CertificatesStatus] = certsStatus
	return status
}

// parse the requested keyvault cert object from the input attributes
func getKeyvaultRequestObj(attrib map[string]string) ([]types.KeyVaultCertificate, error) {
	keyVaultCerts := []types.KeyVaultCertificate{}

	certificatesStrings := types.GetCertificates(attrib)
	if certificatesStrings == "" {
		return nil, fmt.Errorf("certificates is not set")
	}

	logrus.Debugf("certificates string defined in ratify certStore class, certificates %v", certificatesStrings)

	objects, err := types.GetCertificatesArray(certificatesStrings)
	if err != nil {
		return nil, fmt.Errorf("failed to yaml unmarshal objects, error: %w", err)
	}
	logrus.Debugf("unmarshaled objects yaml, objectsArray %v", objects.Array)

	for i, object := range objects.Array {
		var keyVaultCert types.KeyVaultCertificate
		if err = yaml.Unmarshal([]byte(object), &keyVaultCert); err != nil {
			return nil, fmt.Errorf("unmarshal failed for keyVaultCerts at index %d, error: %w", i, err)
		}
		// remove whitespace from all fields in keyVaultCert
		formatKeyVaultCertificate(&keyVaultCert)

		keyVaultCerts = append(keyVaultCerts, keyVaultCert)
	}

	logrus.Debugf("unmarshaled %v key vault objects, keyVaultObjects: %v", len(keyVaultCerts), keyVaultCerts)
	return keyVaultCerts, nil
}

// return a certificate status object that consist of the cert name, version and last refreshed time
func getCertStatusProperty(certificateName, version, lastRefreshed string) map[string]string {
	certProperty := map[string]string{}
	certProperty[types.CertificateName] = certificateName
	certProperty[types.CertificateVersion] = version
	certProperty[types.CertificateLastRefreshed] = lastRefreshed
	return certProperty
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

// parseAzureEnvironment returns azure environment by name
func parseAzureEnvironment(cloudName string) (*azure.Environment, error) {
	var env azure.Environment
	var err error
	if cloudName == "" {
		env = azure.PublicCloud
	} else {
		env, err = azure.EnvironmentFromName(cloudName)
	}
	return &env, err
}

func initializeKvClient(ctx context.Context, keyVaultEndpoint, tenantID, clientID string) (*kv.BaseClient, error) {
	kvClient := kv.New()
	kvEndpoint := strings.TrimSuffix(keyVaultEndpoint, "/")

	err := kvClient.AddToUserAgent("ratify")
	if err != nil {
		return nil, fmt.Errorf("failed to add user agent to keyvault client, error: %w", err)
	}

	kvClient.Authorizer, err = getAuthorizerForWorkloadIdentity(ctx, tenantID, clientID, kvEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get authorizer for keyvault client, error: %w", err)
	}
	return &kvClient, nil
}

// Parse the secret bundle and return an array of certificates
// In a certificate chain scenario, all certificate including root and leaf will be returned
func getCertsFromSecretBundle(secretBundle kv.SecretBundle, certName string) ([]*x509.Certificate, []map[string]string, error) {
	// validation
	if secretBundle.ContentType == nil || secretBundle.Value == nil || secretBundle.ID == nil {
		return nil, nil, errors.Errorf("invalid secret bundle, ContentType, value, and ID must not be nil")
	}

	// This aligns with notation akv implementation
	// akv plugin supports both PKCS12 and PEM. https://github.com/Azure/notation-azure-kv/blob/558e7345ef8318783530de6a7a0a8420b9214ba8/Notation.Plugin.AzureKeyVault/KeyVault/KeyVaultClient.cs#L192
	if *secretBundle.ContentType != PKCS12ContentType &&
		*secretBundle.ContentType != PEMContentType {
		return nil, nil, errors.Errorf("Unsupported secret content type %s, supported type are %s and %s", *secretBundle.ContentType, PKCS12ContentType, PEMContentType)
	}

	if secretBundle.Value == nil {
		return nil, nil, errors.Errorf("azure keyvualt certificate provider: secret value is nil")
	}

	version := getObjectVersion(*secretBundle.ID)

	results := []*x509.Certificate{}
	certsStatus := []map[string]string{}
	lastRefreshed := time.Now().Format(time.RFC3339)

	block, rest := pem.Decode([]byte(*secretBundle.Value))
	for block != nil {
		switch block.Type {
		case "PRIVATE KEY":
			logrus.Warn(" azure keyvualt certificate provider: private key skipped. Please configure your private key to be non exportable.")
		case "CERTIFICATE":
			var pemData []byte
			pemData = append(pemData, pem.EncodeToMemory(block)...)
			decodedCerts, err := certificateprovider.DecodeCertificates(pemData)
			if err != nil {
				return nil, nil, fmt.Errorf("azure keyvualt certificate provider: failed to decode certificate %s, error: %w", certName, err)
			}
			for _, cert := range decodedCerts {
				results = append(results, cert)
				certProperty := getCertStatusProperty(certName, version, lastRefreshed)
				certsStatus = append(certsStatus, certProperty)
				logrus.Debugf("azurekeyvault cert provider: cert '%v', version '%v' added", certName, version)
			}
		default:
			logrus.Warn("azure keyvualt certificate provider detected unknown block type", block.Type)
		}

		block, rest = pem.Decode(rest)
		if block == nil && len(rest) > 0 {
			return nil, nil, errors.Errorf("azure keyvualt certificate provider error: block is nil and remaining block to parse > 0")
		}
	}
	return results, certsStatus, nil
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
