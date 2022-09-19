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
	"fmt"
	"strings"
)

const (
	dockerTokenLoginUsernameGUID string = "00000000-0000-0000-0000-000000000000"
	AADResourcePublicCloud       string = "https://containerregistry.azure.net/.default"
	AADResourceUSGovernmentCloud string = "https://containerregistry.azure.us/.default"
	AADResourceChinaCloud        string = "https://containerregistry.azure.cn/.default"
)

func getAADResource(cloud string) (string, error) {
	if strings.EqualFold(cloud, "") || strings.EqualFold(cloud, "AzurePublicCloud") || strings.EqualFold(cloud, "AzureCloud") {
		return AADResourcePublicCloud, nil
	}
	if strings.EqualFold(cloud, "AzureUSGovernment") || strings.EqualFold(cloud, "AzureUSGovernmentCloud") {
		return AADResourceUSGovernmentCloud, nil
	}
	if strings.EqualFold(cloud, "AzureChinaCloud") {
		return AADResourceChinaCloud, nil
	}
	return "", fmt.Errorf("unsupported cloud %s", cloud)
}
