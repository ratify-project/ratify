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

package awsauth

import "testing"

const (
	image    = "123456789012.dkr.ecr.us-east-2.amazonaws.com/pause:3.1"
	registry = "123456789012.dkr.ecr.us-east-2.amazonaws.com"
	region   = "us-east-2"
)

func TestRegistryFromImage_ReturnsExpected(t *testing.T) {
	reg, err := RegistryFromImage(image)

	if reg == "" || err != nil {
		t.Fatalf("registry parsing failed, expected registry but returned error %v", err)
	}

	if reg != registry {
		t.Fatalf("incorrect registry returned, expected %s, but received %s", registry, reg)
	}
}

func TestRegionFromRegistry_ReturnsExpected(t *testing.T) {
	reg := RegionFromRegistry(registry)
	if reg != region {
		t.Fatalf("incorrect region returned, expected %s, but received %s", region, reg)
	}
}
