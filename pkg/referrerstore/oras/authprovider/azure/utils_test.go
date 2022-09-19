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

package azure

import (
	"testing"
)

// Verifies that Enabled checks if tenantID is empty or AAD token is empty
func TestGetAADResource_ExpectedResults(t *testing.T) {
	cloud1 := ""
	cloud2 := "AzureStackCloud"
	cloud3 := "AzureChinaCloud"
	cloud4 := "azureusGovernment"
	cloud5 := "AzurePublicCloud"

	res, err := getAADResource(cloud1)
	if res != AADResourcePublicCloud {
		t.Fatalf("Get AAD Resource Endpoint failed with input %s, expected: %s, get %s", cloud1, AADResourcePublicCloud, res)
	}

	res, err = getAADResource(cloud2)
	if err == nil {
		t.Fatalf("Get AAD Resource Endpoint should failed with input %s", cloud2)
	}

	res, err = getAADResource(cloud3)
	if res != AADResourceChinaCloud {
		t.Fatalf("Get AAD Resource Endpoint failed with input %s, expected: %s, get %s", cloud3, AADResourceChinaCloud, res)
	}

	res, err = getAADResource(cloud4)
	if res != AADResourceUSGovernmentCloud {
		t.Fatalf("Get AAD Resource Endpoint failed with input %s, expected: %s, get %s", cloud4, AADResourceUSGovernmentCloud, res)
	}

	res, err = getAADResource(cloud5)
	if res != AADResourcePublicCloud {
		t.Fatalf("Get AAD Resource Endpoint failed with input %s, expected: %s, get %s", cloud5, AADResourcePublicCloud, res)
	}
}
