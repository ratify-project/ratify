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
package types

// This class is based on implementation from azure secret store csi provider
// Source: https://github.com/Azure/secrets-store-csi-driver-provider-azure/tree/release-1.4/pkg/provider
import (
	"reflect"
	"testing"
)

func TestGetKeyVaultName(t *testing.T) {
	tests := []struct {
		name       string
		parameters map[string]string
		expected   string
	}{
		{
			name: "empty",
			parameters: map[string]string{
				KeyVaultNameParameter: "",
			},
			expected: "",
		},
		{
			name: "not empty",
			parameters: map[string]string{
				KeyVaultNameParameter: "test",
			},
			expected: "test",
		},
		{
			name: "trim spaces",
			parameters: map[string]string{
				KeyVaultNameParameter: " test ",
			},
			expected: "test",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := GetKeyVaultName(test.parameters)
			if actual != test.expected {
				t.Errorf("GetKeyVaultName() = %v, expected %v", actual, test.expected)
			}
		})
	}
}

func TestGetCloudName(t *testing.T) {
	tests := []struct {
		name       string
		parameters map[string]string
		expected   string
	}{
		{
			name: "empty",
			parameters: map[string]string{
				CloudNameParameter: "",
			},
			expected: "",
		},
		{
			name: "not empty",
			parameters: map[string]string{
				CloudNameParameter: "test",
			},
			expected: "test",
		},
		{
			name: "trim spaces",
			parameters: map[string]string{
				CloudNameParameter: " test ",
			},
			expected: "test",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := GetCloudName(test.parameters)
			if actual != test.expected {
				t.Errorf("GetCloudName() = %v, expected %v", actual, test.expected)
			}
		})
	}
}

func TestGetTenantID(t *testing.T) {
	tests := []struct {
		name       string
		parameters map[string]string
		expected   string
	}{
		{
			name: "empty",
			parameters: map[string]string{
				TenantIDParameter: "",
			},
			expected: "",
		},
		{
			name: "not empty",
			parameters: map[string]string{
				TenantIDParameter: "test",
			},
			expected: "test",
		},
		{
			name: "trim spaces",
			parameters: map[string]string{
				TenantIDParameter: " test ",
			},
			expected: "test",
		},
		{
			name: "new tenantID parameter",
			parameters: map[string]string{
				"tenantID": "test",
			},
			expected: "test",
		},
		{
			name: "new tenantID parameter with spaces",
			parameters: map[string]string{
				"tenantID": " test ",
			},
			expected: "test",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := GetTenantID(test.parameters)
			if actual != test.expected {
				t.Errorf("GetTenantID() = %v, expected %v", actual, test.expected)
			}
		})
	}
}

func TestGetClientID(t *testing.T) {
	tests := []struct {
		name       string
		parameters map[string]string
		expected   string
	}{
		{
			name: "empty",
			parameters: map[string]string{
				"clientID": "",
			},
			expected: "",
		},
		{
			name: "not empty",
			parameters: map[string]string{
				"clientID": "test",
			},
			expected: "test",
		},
		{
			name: "trim spaces",
			parameters: map[string]string{
				"clientID": " test ",
			},
			expected: "test",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := GetClientID(test.parameters)
			if actual != test.expected {
				t.Errorf("GetClientID() = %v, expected %v", actual, test.expected)
			}
		})
	}
}

func TestGetObjects(t *testing.T) {
	tests := []struct {
		name       string
		parameters map[string]string
		expected   string
	}{
		{
			name: "empty",
			parameters: map[string]string{
				CertificatesParameter: "",
			},
			expected: "",
		},
		{
			name: "not empty",
			parameters: map[string]string{
				CertificatesParameter: "test",
			},
			expected: "test",
		},
		{
			name: "trim spaces",
			parameters: map[string]string{
				CertificatesParameter: " test ",
			},
			expected: "test",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := GetCertificates(test.parameters)
			if actual != test.expected {
				t.Errorf("GetObjects() = %v, expected %v", actual, test.expected)
			}
		})
	}
}

func TestGetObjectsArray(t *testing.T) {
	tests := []struct {
		name     string
		objects  string
		expected StringArray
	}{
		{
			name:     "empty",
			objects:  "",
			expected: StringArray{},
		},
		{
			name:    "valid yaml",
			objects: "array:\n- |\n  filePermission: \"\"\n  objectAlias: \"\"\n  objectEncoding: \"\"\n  objectFormat: \"\"\n  objectName: secret1\n  objectType: cert\n  objectVersion: \"\"\n- |\n  filePermission: \"\"\n  objectAlias: \"\"\n  objectEncoding: \"\"\n  objectFormat: \"\"\n  objectName: secret2\n  objectType: cert\n  objectVersion: \"\"\n",
			expected: StringArray{
				Array: []string{
					"filePermission: \"\"\nobjectAlias: \"\"\nobjectEncoding: \"\"\nobjectFormat: \"\"\nobjectName: secret1\nobjectType: cert\nobjectVersion: \"\"\n",
					"filePermission: \"\"\nobjectAlias: \"\"\nobjectEncoding: \"\"\nobjectFormat: \"\"\nobjectName: secret2\nobjectType: cert\nobjectVersion: \"\"\n",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := GetCertificatesArray(test.objects)
			if err != nil {
				t.Errorf("GetObjectsArray() error = %v", err)
			}
			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("GetObjectsArray() = %v, expected %v", actual, test.expected)
			}
		})
	}
}

func TestGetObjectsArrayError(t *testing.T) {
	objects := "invalid"
	if _, err := GetCertificatesArray(objects); err == nil {
		t.Errorf("GetObjectsArray() error is nil, expected error")
	}
}

func TestIsSyncingSingleVersion(t *testing.T) {
	tests := []struct {
		name     string
		object   KeyVaultCertificate
		expected bool
	}{
		{
			name:     "object version history uninitialized",
			object:   KeyVaultCertificate{},
			expected: true,
		},
		{
			name: "object version history set to 0",
			object: KeyVaultCertificate{
				CertificateVersionHistory: 0,
			},
			expected: true,
		},
		{
			name: "object version history set to 1",
			object: KeyVaultCertificate{
				CertificateVersionHistory: 1,
			},
			expected: true,
		},
		{
			name: "object version history set higher than 1",
			object: KeyVaultCertificate{
				CertificateVersionHistory: 4,
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := test.object.IsSyncingSingleVersion()
			if actual != test.expected {
				t.Errorf("IsSyncingSingleVersion() = %v, expected %v", actual, test.expected)
			}
		})
	}
}
