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
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/deislabs/ratify/pkg/certificateprovider/azurekeyvault/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestParseAzureEnvironment(t *testing.T) {
	envNamesArray := []string{"AZURECHINACLOUD", "AZUREGERMANCLOUD", "AZUREPUBLICCLOUD", "AZUREUSGOVERNMENTCLOUD", ""}
	for _, envName := range envNamesArray {
		azureEnv, err := parseAzureEnvironment(envName)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if strings.EqualFold(envName, "") && !strings.EqualFold(azureEnv.Name, "AZUREPUBLICCLOUD") {
			t.Fatalf("string doesn't match, expected AZUREPUBLICCLOUD, got %s", azureEnv.Name)
		} else if !strings.EqualFold(envName, "") && !strings.EqualFold(envName, azureEnv.Name) {
			t.Fatalf("string doesn't match, expected %s, got %s", envName, azureEnv.Name)
		}
	}

	wrongEnvName := "AZUREWRONGCLOUD"
	_, err := parseAzureEnvironment(wrongEnvName)
	if err == nil {
		t.Fatalf("expected error for wrong azure environment name")
	}
}

func TestFormatKeyVaultCertificate(t *testing.T) {
	cases := []struct {
		desc                   string
		keyVaultObject         types.KeyVaultCertificate
		expectedKeyVaultObject types.KeyVaultCertificate
	}{
		{
			desc: "leading and trailing whitespace trimmed from all fields",
			keyVaultObject: types.KeyVaultCertificate{
				CertificateName:    "cert1     ",
				CertificateVersion: "",
			},
			expectedKeyVaultObject: types.KeyVaultCertificate{
				CertificateName:    "cert1",
				CertificateVersion: "",
			},
		},
		{
			desc: "no data loss for already sanitized object",
			keyVaultObject: types.KeyVaultCertificate{
				CertificateName:    "cert1",
				CertificateVersion: "version1",
			},
			expectedKeyVaultObject: types.KeyVaultCertificate{
				CertificateName:    "cert1",
				CertificateVersion: "version1",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			formatKeyVaultCertificate(&tc.keyVaultObject)
			if !reflect.DeepEqual(tc.keyVaultObject, tc.expectedKeyVaultObject) {
				t.Fatalf("expected: %+v, but got: %+v", tc.expectedKeyVaultObject, tc.keyVaultObject)
			}
		})
	}
}

func SkipTestInitializeKVClient(t *testing.T) {
	testEnvs := []azure.Environment{
		azure.PublicCloud,
		azure.GermanCloud,
		azure.ChinaCloud,
		azure.USGovernmentCloud,
	}

	for i := range testEnvs {
		kvBaseClient, err := initializeKvClient(context.TODO(), testEnvs[i].KeyVaultEndpoint, "", "")
		assert.NoError(t, err)
		assert.NotNil(t, kvBaseClient)
		assert.NotNil(t, kvBaseClient.Authorizer)
		assert.Contains(t, kvBaseClient.UserAgent, "ratify")
	}
}

func TestGetCertificates(t *testing.T) {
	cases := []struct {
		desc        string
		parameters  map[string]string
		secrets     map[string]string
		expectedErr bool
	}{
		{
			desc:        "keyvault uri not provided",
			parameters:  map[string]string{},
			expectedErr: true,
		},
		{
			desc: "tenantID not provided",
			parameters: map[string]string{
				"vaultUri": "https://testkv.vault.azure.net/",
			},
			expectedErr: true,
		},
		{
			desc: "clientID not provided",
			parameters: map[string]string{
				"vaultUri": "https://testkv.vault.azure.net/",
				"tenantID": "tid",
			},
			expectedErr: true,
		},
		{
			desc: "invalid cloud name",
			parameters: map[string]string{
				"vaultUri":  "https://testkv.vault.azure.net/",
				"tenantID":  "tid",
				"cloudName": "AzureCloud",
			},
			expectedErr: true,
		},
		{
			desc: "certificates array not set",
			parameters: map[string]string{
				"vaultUri":             "https://testkv.vault.azure.net/",
				"tenantID":             "tid",
				"useVMManagedIdentity": "true",
			},
			expectedErr: true,
		},
		{
			desc: "certificates not configured as an array",
			parameters: map[string]string{
				"vaultUri":             "https://testkv.vault.azure.net/",
				"tenantID":             "tid",
				"useVMManagedIdentity": "true",
				"certificates": `
        - |
          CertificateName: cert1
          CertificateVersion: ""`,
			},
			expectedErr: true,
		},
		{
			desc: "certificates array is empty",
			parameters: map[string]string{
				"vaultUri": "https://testkv.vault.azure.net/",
				"tenantID": "tid",
				"clientID": "clientid",
				"certificates": `
      array:`,
			},
			expectedErr: true,
		},
		{
			desc: "invalid object format",
			parameters: map[string]string{
				"vaultUri": "https://testkv.vault.azure.net/",
				"tenantID": "tid",
				"clientID": "clientid",
				"certificates": `
      array:
        - |
          CertificateName: cert1
          CertificateVersion: ""`,
			},
			expectedErr: true,
		},
		{
			desc: "error fetching from keyvault",
			parameters: map[string]string{
				"vaultUri": "https://testkv.vault.azure.net/",
				"tenantID": "tid",
				"certificates": `
      array:
        - |
          CertificateName: cert1
          CertificateVersion: ""`,
			},
			expectedErr: true,
		},
	}

	provider := Create()

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			certs, certStatus, err := provider.GetCertificates(context.TODO(), tc.parameters)
			if tc.expectedErr {
				assert.NotNil(t, err)
				assert.Nil(t, certs)
				assert.Nil(t, certStatus)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestGetCertStatusMap(t *testing.T) {
	certsStatus := []map[string]string{}
	certsStatus = append(certsStatus, map[string]string{
		"CertName":    "Cert1",
		"CertVersion": "VersionABC",
	})
	certsStatus = append(certsStatus, map[string]string{
		"CertName":    "Cert2",
		"CertVersion": "VersionEDF",
	})

	actual := getCertStatusMap(certsStatus)
	assert.NotNil(t, actual[types.CertificatesStatus])
}

func TestGetObjectVersion(t *testing.T) {
	id := "https://kindkv.vault.azure.net/secrets/cert1/c55925c29c6743dcb9bb4bf091be03b0"
	expectedVersion := "c55925c29c6743dcb9bb4bf091be03b0"
	actual := getObjectVersion(id)
	assert.Equal(t, expectedVersion, actual)
}

func TestGetCertStatusProperty(t *testing.T) {
	timeNow := time.Now().String()
	certName := "certName"
	certVersion := "versionABC"

	status := getCertStatusProperty(certName, certVersion, timeNow)
	assert.Equal(t, certName, status[types.CertificateName])
	assert.Equal(t, timeNow, status[types.CertificateLastRefreshed])
	assert.Equal(t, certVersion, status[types.CertificateVersion])
}

func TestGetKeyvaultRequestObj(t *testing.T) {
	attrib := map[string]string{}
	attrib["vaultURI"] = "https://testvault.vault.azure.net/"
	attrib["clientID"] = "TestClient"
	attrib["cloudName"] = "TestCloud"
	attrib["tenantID"] = "TestIDABC"
	attrib["certificates"] = "array:\n- |\n  certificateName: wabbit-networks-io  \n  certificateVersion: \"testversion\"\n"

	result, err := getKeyvaultRequestObj(attrib)

	if err != nil {
		logrus.Infof("err %v", err)
	}

	assert.Equal(t, "wabbit-networks-io", result[0].CertificateName)
	assert.Equal(t, "testversion", result[0].CertificateVersion)
}

func TestGetKeyvaultRequestObj_error(t *testing.T) {
	cases := []struct {
		desc        string
		attrib      map[string]string
		expectedErr bool
	}{
		{
			desc: "certificates is not set",
			attrib: map[string]string{
				"certificates": "",
			},
			expectedErr: true,
		},
		{
			desc: "unexpected certificate string",
			attrib: map[string]string{
				"certificates": "certificateName: wabbit-networks-io  \n  certificateVersion: \"testversion\"\n",
			},
			expectedErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			_, err := getKeyvaultRequestObj(tc.attrib)
			if tc.expectedErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
