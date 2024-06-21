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
	"crypto"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v3"
	re "github.com/ratify-project/ratify/errors"
	"github.com/ratify-project/ratify/internal/logger"
	"github.com/ratify-project/ratify/pkg/keymanagementprovider"
	"github.com/ratify-project/ratify/pkg/keymanagementprovider/azurekeyvault/types"
	"github.com/ratify-project/ratify/pkg/keymanagementprovider/config"
	"github.com/ratify-project/ratify/pkg/keymanagementprovider/factory"
	"github.com/ratify-project/ratify/pkg/metrics"
	"golang.org/x/crypto/pkcs12"

	kv "github.com/Azure/azure-sdk-for-go/services/keyvault/v7.1/keyvault"
	"github.com/Azure/go-autorest/autorest/azure"
)

const (
	ProviderName      string = "azurekeyvault"
	PKCS12ContentType string = "application/x-pkcs12"
	PEMContentType    string = "application/x-pem-file"
)

var logOpt = logger.Option{
	ComponentType: logger.KeyManagementProvider,
}

type AKVKeyManagementProviderConfig struct {
	Type         string                `json:"type"`
	VaultURI     string                `json:"vaultURI"`
	TenantID     string                `json:"tenantID"`
	ClientID     string                `json:"clientID"`
	CloudName    string                `json:"cloudName,omitempty"`
	Certificates []types.KeyVaultValue `json:"certificates,omitempty"`
	Keys         []types.KeyVaultValue `json:"keys,omitempty"`
}

type akvKMProvider struct {
	provider     string
	vaultURI     string
	tenantID     string
	clientID     string
	cloudName    string
	certificates []types.KeyVaultValue
	keys         []types.KeyVaultValue
	cloudEnv     *azure.Environment
	kvClient     *kv.BaseClient
}
type akvKMProviderFactory struct{}

// initKVClient is a function to initialize the keyvault client
// used for mocking purposes
var initKVClient = initializeKvClient

// init calls to register the provider
func init() {
	factory.Register(ProviderName, &akvKMProviderFactory{})
}

// Create creates a new instance of the provider after marshalling and validating the configuration
func (f *akvKMProviderFactory) Create(_ string, keyManagementProviderConfig config.KeyManagementProviderConfig, _ string) (keymanagementprovider.KeyManagementProvider, error) {
	conf := AKVKeyManagementProviderConfig{}

	keyManagementProviderConfigBytes, err := json.Marshal(keyManagementProviderConfig)
	if err != nil {
		return nil, re.ErrorCodeConfigInvalid.WithError(err).WithComponentType(re.KeyManagementProvider)
	}

	if err := json.Unmarshal(keyManagementProviderConfigBytes, &conf); err != nil {
		return nil, re.ErrorCodeConfigInvalid.NewError(re.KeyManagementProvider, "", re.EmptyLink, err, "failed to parse AKV key management provider configuration", re.HideStackTrace)
	}

	azureCloudEnv, err := parseAzureEnvironment(conf.CloudName)
	if err != nil {
		return nil, re.ErrorCodeConfigInvalid.NewError(re.KeyManagementProvider, ProviderName, re.EmptyLink, nil, fmt.Sprintf("cloudName %s is not valid", conf.CloudName), re.HideStackTrace)
	}

	if len(conf.Certificates) == 0 && len(conf.Keys) == 0 {
		return nil, re.ErrorCodeConfigInvalid.NewError(re.KeyManagementProvider, ProviderName, re.EmptyLink, nil, "no keyvault certificates or keys configured", re.HideStackTrace)
	}

	provider := &akvKMProvider{
		provider:     ProviderName,
		vaultURI:     strings.TrimSpace(conf.VaultURI),
		tenantID:     strings.TrimSpace(conf.TenantID),
		clientID:     strings.TrimSpace(conf.ClientID),
		cloudName:    strings.TrimSpace(conf.CloudName),
		certificates: conf.Certificates,
		keys:         conf.Keys,
		cloudEnv:     azureCloudEnv,
	}
	if err := provider.validate(); err != nil {
		return nil, err
	}

	logger.GetLogger(context.Background(), logOpt).Debugf("vaultURI %s", provider.vaultURI)

	kvClient, err := initKVClient(context.Background(), provider.cloudEnv.KeyVaultEndpoint, provider.tenantID, provider.clientID)
	if err != nil {
		return nil, re.ErrorCodePluginInitFailure.NewError(re.KeyManagementProvider, ProviderName, re.AKVLink, err, "failed to create keyvault client", re.HideStackTrace)
	}
	provider.kvClient = kvClient

	return provider, nil
}

// GetCertificates returns an array of certificates based on certificate properties defined in config
// get certificate retrieve the entire cert chain using getSecret API call
func (s *akvKMProvider) GetCertificates(ctx context.Context) (map[keymanagementprovider.KMPMapKey][]*x509.Certificate, keymanagementprovider.KeyManagementProviderStatus, error) {
	certsMap := map[keymanagementprovider.KMPMapKey][]*x509.Certificate{}
	certsStatus := []map[string]string{}
	for _, keyVaultCert := range s.certificates {
		logger.GetLogger(ctx, logOpt).Debugf("fetching secret from key vault, certName %v,  keyvault %v", keyVaultCert.Name, s.vaultURI)

		// fetch the object from Key Vault
		// GetSecret is required so we can fetch the entire cert chain. See issue https://github.com/ratify-project/ratify/issues/695 for details
		startTime := time.Now()
		secretBundle, err := s.kvClient.GetSecret(ctx, s.vaultURI, keyVaultCert.Name, keyVaultCert.Version)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get secret objectName:%s, objectVersion:%s, error: %w", keyVaultCert.Name, keyVaultCert.Version, err)
		}

		certResult, certProperty, err := getCertsFromSecretBundle(ctx, secretBundle, keyVaultCert.Name)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get certificates from secret bundle:%w", err)
		}

		metrics.ReportAKVCertificateDuration(ctx, time.Since(startTime).Milliseconds(), keyVaultCert.Name)
		certsStatus = append(certsStatus, certProperty...)
		certMapKey := keymanagementprovider.KMPMapKey{Name: keyVaultCert.Name, Version: keyVaultCert.Version}
		certsMap[certMapKey] = certResult
	}

	return certsMap, getStatusMap(certsStatus, types.CertificatesStatus), nil
}

// GetKeys returns an array of keys based on key properties defined in config
func (s *akvKMProvider) GetKeys(ctx context.Context) (map[keymanagementprovider.KMPMapKey]crypto.PublicKey, keymanagementprovider.KeyManagementProviderStatus, error) {
	keysMap := map[keymanagementprovider.KMPMapKey]crypto.PublicKey{}
	keysStatus := []map[string]string{}

	for _, keyVaultKey := range s.keys {
		logger.GetLogger(ctx, logOpt).Debugf("fetching key from key vault, keyName %v,  keyvault %v", keyVaultKey.Name, s.vaultURI)

		// fetch the key object from Key Vault
		startTime := time.Now()
		keyBundle, err := s.kvClient.GetKey(ctx, s.vaultURI, keyVaultKey.Name, keyVaultKey.Version)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get key objectName:%s, objectVersion:%s, error: %w", keyVaultKey.Name, keyVaultKey.Version, err)
		}

		if keyBundle.Attributes != nil && keyBundle.Attributes.Enabled != nil && !*keyBundle.Attributes.Enabled {
			return nil, nil, fmt.Errorf("key %s version %s is disabled. please re-enable in azure key vault or remove reference to this key", keyVaultKey.Name, keyVaultKey.Version)
		}

		publicKey, err := getKeyFromKeyBundle(keyBundle)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get key from key bundle:%w", err)
		}
		keysMap[keymanagementprovider.KMPMapKey{Name: keyVaultKey.Name, Version: keyVaultKey.Version}] = publicKey
		metrics.ReportAKVCertificateDuration(ctx, time.Since(startTime).Milliseconds(), keyVaultKey.Name)
		properties := getStatusProperty(keyVaultKey.Name, keyVaultKey.Version, time.Now().Format(time.RFC3339))
		keysStatus = append(keysStatus, properties)
	}

	return keysMap, getStatusMap(keysStatus, types.KeysStatus), nil
}

// azure keyvault provider certificate/key status is a map from "certificates" key or "keys" key to an array of key management provider status
func getStatusMap(statusMap []map[string]string, contentType string) keymanagementprovider.KeyManagementProviderStatus {
	status := keymanagementprovider.KeyManagementProviderStatus{}
	status[contentType] = statusMap
	return status
}

// return a status object that consist of the cert/key name, version and last refreshed time
func getStatusProperty(name, version, lastRefreshed string) map[string]string {
	properties := map[string]string{}
	properties[types.StatusName] = name
	properties[types.StatusVersion] = version
	properties[types.StatusLastRefreshed] = lastRefreshed
	return properties
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
		return nil, re.ErrorCodeConfigInvalid.NewError(re.KeyManagementProvider, ProviderName, re.AKVLink, err, "failed to add user agent to keyvault client", re.PrintStackTrace)
	}

	kvClient.Authorizer, err = getAuthorizerForWorkloadIdentity(ctx, tenantID, clientID, kvEndpoint)
	if err != nil {
		return nil, re.ErrorCodeAuthDenied.NewError(re.KeyManagementProvider, ProviderName, re.AKVLink, err, "failed to get authorizer for keyvault client", re.PrintStackTrace)
	}
	return &kvClient, nil
}

// Parse the secret bundle and return an array of certificates
// In a certificate chain scenario, all certificates from root to leaf will be returned
func getCertsFromSecretBundle(ctx context.Context, secretBundle kv.SecretBundle, certName string) ([]*x509.Certificate, []map[string]string, error) {
	if secretBundle.ContentType == nil || secretBundle.Value == nil || secretBundle.ID == nil {
		return nil, nil, re.ErrorCodeCertInvalid.NewError(re.KeyManagementProvider, ProviderName, re.EmptyLink, nil, "found invalid secret bundle for certificate  %s, contentType, value, and id must not be nil", re.HideStackTrace)
	}

	version := getObjectVersion(*secretBundle.ID)

	// This aligns with notation akv implementation
	// akv plugin supports both PKCS12 and PEM. https://github.com/Azure/notation-azure-kv/blob/558e7345ef8318783530de6a7a0a8420b9214ba8/Notation.Plugin.AzureKeyVault/KeyVault/KeyVaultClient.cs#L192
	if *secretBundle.ContentType != PKCS12ContentType &&
		*secretBundle.ContentType != PEMContentType {
		return nil, nil, re.ErrorCodeCertInvalid.NewError(re.KeyManagementProvider, ProviderName, re.EmptyLink, nil, fmt.Sprintf("certificate %s version %s, unsupported secret content type %s, supported type are %s and %s", certName, version, *secretBundle.ContentType, PKCS12ContentType, PEMContentType), re.HideStackTrace)
	}

	results := []*x509.Certificate{}
	certsStatus := []map[string]string{}
	lastRefreshed := time.Now().Format(time.RFC3339)

	data := []byte(*secretBundle.Value)

	if *secretBundle.ContentType == PKCS12ContentType {
		p12, err := base64.StdEncoding.DecodeString(*secretBundle.Value)
		if err != nil {
			return nil, nil, re.ErrorCodeCertInvalid.NewError(re.KeyManagementProvider, ProviderName, re.EmptyLink, err, fmt.Sprintf("azure keyvault key management provider: failed to decode PKCS12 Value. Certificate %s, version %s", certName, version), re.HideStackTrace)
		}

		blocks, err := pkcs12.ToPEM(p12, "")
		if err != nil {
			return nil, nil, re.ErrorCodeCertInvalid.NewError(re.KeyManagementProvider, ProviderName, re.EmptyLink, err, fmt.Sprintf("azure keyvault key management provider: failed to convert PKCS12 Value to PEM. Certificate %s, version %s", certName, version), re.HideStackTrace)
		}

		var pemData []byte
		for _, b := range blocks {
			pemData = append(pemData, pem.EncodeToMemory(b)...)
		}
		data = pemData
	}

	block, rest := pem.Decode(data)

	for block != nil {
		switch block.Type {
		case "PRIVATE KEY":
			logger.GetLogger(ctx, logOpt).Warnf("azure keyvault key management provider: certificate %s, version %s private key skipped. Please see doc to learn how to create a new certificate in keyvault with non exportable keys. https://learn.microsoft.com/en-us/azure/key-vault/certificates/how-to-export-certificate?tabs=azure-cli#exportable-and-non-exportable-keys", certName, version)
		case "CERTIFICATE":
			var pemData []byte
			pemData = append(pemData, pem.EncodeToMemory(block)...)
			decodedCerts, err := keymanagementprovider.DecodeCertificates(pemData)
			if err != nil {
				return nil, nil, re.ErrorCodeCertInvalid.NewError(re.KeyManagementProvider, ProviderName, re.EmptyLink, err, fmt.Sprintf("azure keyvault key management provider: failed to decode Certificate %s, version %s", certName, version), re.HideStackTrace)
			}
			for _, cert := range decodedCerts {
				results = append(results, cert)
				certProperty := getStatusProperty(certName, version, lastRefreshed)
				certsStatus = append(certsStatus, certProperty)
			}
		default:
			logger.GetLogger(ctx, logOpt).Warnf("certificate '%s', version '%s': azure keyvault key management provider detected unknown block type %s", certName, version, block.Type)
		}

		block, rest = pem.Decode(rest)
		if block == nil && len(rest) > 0 {
			return nil, nil, re.ErrorCodeCertInvalid.NewError(re.KeyManagementProvider, ProviderName, re.EmptyLink, nil, fmt.Sprintf("certificate '%s', version '%s': azure keyvault key management provider error, block is nil and remaining block to parse > 0", certName, version), re.HideStackTrace)
		}
	}
	logger.GetLogger(ctx, logOpt).Debugf("azurekeyvault certprovider getCertsFromSecretBundle: %v certificates parsed, Certificate '%s', version '%s'", len(results), certName, version)
	return results, certsStatus, nil
}

// Based on https://github.com/sigstore/sigstore/blob/8b208f7d608b80a7982b2a66358b8333b1eec542/pkg/signature/kms/azure/client.go#L258
func getKeyFromKeyBundle(keyBundle kv.KeyBundle) (crypto.PublicKey, error) {
	webKey := keyBundle.Key
	if webKey == nil {
		return nil, re.ErrorCodeKeyInvalid.NewError(re.KeyManagementProvider, ProviderName, re.EmptyLink, nil, "found invalid key bundle, key must not be nil", re.HideStackTrace)
	}

	keyType := webKey.Kty
	switch keyType {
	case kv.ECHSM:
		webKey.Kty = kv.EC
	case kv.RSAHSM:
		webKey.Kty = kv.RSA
	}

	keyBytes, err := json.Marshal(webKey)
	if err != nil {
		return nil, re.ErrorCodeKeyInvalid.NewError(re.KeyManagementProvider, ProviderName, re.EmptyLink, err, "failed to marshal key", re.HideStackTrace)
	}

	key := jose.JSONWebKey{}
	err = key.UnmarshalJSON(keyBytes)
	if err != nil {
		return nil, re.ErrorCodeKeyInvalid.NewError(re.KeyManagementProvider, ProviderName, re.EmptyLink, err, "failed to unmarshal key into JSON Web Key", re.HideStackTrace)
	}

	return key.Key, nil
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

// validate checks vaultURI, tenantID, clientID are set and all certificates/keys have a name
func (s *akvKMProvider) validate() error {
	if s.vaultURI == "" {
		return re.ErrorCodeConfigInvalid.NewError(re.KeyManagementProvider, ProviderName, re.EmptyLink, nil, "vaultURI is not set", re.HideStackTrace)
	}
	if s.tenantID == "" {
		return re.ErrorCodeConfigInvalid.NewError(re.KeyManagementProvider, ProviderName, re.EmptyLink, nil, "tenantID is not set", re.HideStackTrace)
	}
	if s.clientID == "" {
		return re.ErrorCodeConfigInvalid.NewError(re.KeyManagementProvider, ProviderName, re.EmptyLink, nil, "clientID is not set", re.HideStackTrace)
	}

	// all certificates must have a name
	for i := range s.certificates {
		if s.certificates[i].Name == "" {
			return re.ErrorCodeConfigInvalid.NewError(re.KeyManagementProvider, ProviderName, re.EmptyLink, nil, fmt.Sprintf("name is not set for the %d th certificate", i+1), re.HideStackTrace)
		}
	}

	// all keys must have a name
	for i := range s.keys {
		if s.keys[i].Name == "" {
			return re.ErrorCodeConfigInvalid.NewError(re.KeyManagementProvider, ProviderName, re.EmptyLink, nil, fmt.Sprintf("name is not set for the %d th key", i+1), re.HideStackTrace)
		}
	}

	return nil
}
