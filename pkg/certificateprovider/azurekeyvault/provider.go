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
	"encoding/pem"
	"fmt"
	"reflect"
	"strings"

	"github.com/deislabs/ratify/pkg/certificateprovider/azurekeyvault/types"
	"github.com/deislabs/ratify/pkg/common"

	kv "github.com/Azure/azure-sdk-for-go/services/keyvault/v7.1/keyvault"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type keyvaultObject struct {
	content string
	version string
}

// returns an array of certificates based on certificate properties defined in attrib map
func GetCertificates(ctx context.Context, attrib map[string]string) ([]types.Certificate, error) {
	keyvaultUri := types.GetKeyVaultUri(attrib)
	cloudName := types.GetCloudName(attrib)
	tenantID := types.GetTenantID(attrib)
	workloadIdentityClientID := types.GetClientID(attrib)

	if keyvaultUri == "" {
		return nil, fmt.Errorf("keyvaultUri is not set")
	}
	if tenantID == "" {
		return nil, fmt.Errorf("tenantID is not set")
	}
	if workloadIdentityClientID == "" {
		return nil, fmt.Errorf("clientID is not set")
	}

	azureCloudEnv, err := parseAzureEnvironment(cloudName)
	if err != nil {
		return nil, fmt.Errorf("cloudName %s is not valid, error: %w", cloudName, err)
	}

	// 1. cleaning up keyvault objects definition
	certificatesStrings := types.GetCertificates(attrib)
	if certificatesStrings == "" {
		return nil, fmt.Errorf("certificates is not set")
	}

	common.LogDebug("certificates string defined in ratify certStore class, certificates %v", certificatesStrings)

	objects, err := types.GetCertificatesArray(certificatesStrings)
	if err != nil {
		return nil, fmt.Errorf("failed to yaml unmarshal objects, error: %w", err)
	}
	common.LogDebug("unmarshaled objects yaml, objectsArray %v", objects.Array)

	keyVaultCerts := []types.KeyVaultCertificate{}
	for i, object := range objects.Array {
		var keyVaultCert types.KeyVaultCertificate
		if err = yaml.Unmarshal([]byte(object), &keyVaultCert); err != nil {
			return nil, fmt.Errorf("unmarshal failed for keyVaultCerts at index %d, error: %w", i, err)
		}
		// remove whitespace from all fields in keyVaultCert
		formatKeyVaultCertificate(&keyVaultCert)

		keyVaultCerts = append(keyVaultCerts, keyVaultCert)
	}

	common.LogDebug("unmarshaled %v key vault objects, keyVaultObjects: %v", len(keyVaultCerts), keyVaultCerts)

	if len(keyVaultCerts) == 0 {
		return nil, fmt.Errorf("no keyvault certificate configured")
	}

	common.LogDebug("vaultURI %s", keyvaultUri)

	kvClient, err := initializeKvClient(ctx, azureCloudEnv.KeyVaultEndpoint, tenantID, workloadIdentityClientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get keyvault client, error: %w", err)
	}

	certs := []types.Certificate{}
	for _, keyVaultCert := range keyVaultCerts {
		common.LogDebug("fetching object from key vault, certName %v,  keyvault %v", keyVaultCert.CertificateName, keyvaultUri)

		// fetch the object from Key Vault
		result, err := getCertificate(ctx, kvClient, keyvaultUri, keyVaultCert)
		if err != nil {
			return nil, err
		}

		objectContent := []byte(result.content)

		cert := types.Certificate{
			CertificateName: keyVaultCert.CertificateName,
			Content:         objectContent,
			Version:         result.version,
		}

		certs = append(certs, cert)
		common.LogDebug("added certificates %v to response", cert.CertificateName)
	}
	return certs, nil
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

func initializeKvClient(ctx context.Context, keyVaultEndpoint, tenantID, clientId string) (*kv.BaseClient, error) {
	kvClient := kv.New()
	kvEndpoint := strings.TrimSuffix(keyVaultEndpoint, "/")

	err := kvClient.AddToUserAgent("ratify")
	if err != nil {
		return nil, fmt.Errorf("failed to add user agent to keyvault client, error: %w", err)
	}

	kvClient.Authorizer, err = getAuthorizerForWorkloadIdentity(ctx, tenantID, clientId, kvEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get authorizer for keyvault client, error: %w", err)
	}
	return &kvClient, nil
}

// getCertificate retrieves the certificate from the vault
func getCertificate(ctx context.Context, kvClient *kv.BaseClient, vaultURL string, kvObject types.KeyVaultCertificate) (keyvaultObject, error) {
	certbundle, err := kvClient.GetCertificate(ctx, vaultURL, kvObject.CertificateName, kvObject.CertificateVersion)
	if err != nil {
		return keyvaultObject{}, fmt.Errorf("failed to get certificate objectName:%s, objectVersion:%s, error: %w", kvObject.CertificateName, kvObject.CertificateVersion, err)
	}
	if certbundle.Cer == nil {
		return keyvaultObject{}, errors.Errorf("certificate value is nil")
	}
	if certbundle.ID == nil {
		return keyvaultObject{}, errors.Errorf("certificate id is nil")
	}
	version := getObjectVersion(*certbundle.ID)

	certBlock := &pem.Block{
		Type:  types.CertificateType,
		Bytes: *certbundle.Cer,
	}
	var pemData []byte
	pemData = append(pemData, pem.EncodeToMemory(certBlock)...)
	return keyvaultObject{content: string(pemData), version: version}, nil
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
