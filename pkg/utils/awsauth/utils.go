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

import (
	"net/url"
	"strings"
)

// RegionFromImage parses region from image url
func RegionFromImage(image string) (string, error) {
	registry, err := RegistryFromImage(image)
	if err != nil {
		return "", err
	}

	region := RegionFromRegistry(registry)

	return region, nil
}

// RegionFromRegistry parses AWS region ID from registry url
func RegionFromRegistry(registry string) string {
	a := strings.Split(registry, ".")
	if len(a) >= 6 {
		return a[3]
	}
	return ""
}

// RegistryFromImage parses registry host from image url
func RegistryFromImage(image string) (string, error) {
	if strings.Contains(image, "https://") {
		u, err := url.Parse(image)
		if err != nil {
			return "", err
		}
		return u.Host, nil
	}

	return image[:strings.IndexByte(image, '/')], nil
}
