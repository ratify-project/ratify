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

// parseEndpoints checks if the endpoints are valid for auth provider. If no
// endpoints are provided, it defaults to the default ACR endpoint.
// A valid endpoint is either a fully qualified domain name or a wildcard domain
// name folloiwing RFC 1034.
// Valid examples:
// - *.example.com
// - example.com
//
// Invalid examples:
// - *
// - example.*
// - *example.com.*
// - *.
func parseEndpoints(endpoints []string) ([]string, error) {
	if len(endpoints) == 0 {
		return defaultACREndpoints, nil
	}
	for _, endpoint := range endpoints {
		switch strings.Count(endpoint, "*") {
		case 0:
			continue
		case 1:
			if !strings.HasPrefix(endpoint, "*.") {
				return nil, fmt.Errorf("invalid wildcard domain name: %s, it must start with '*.'", endpoint)
			}
			if len(endpoint) < 3 {
				return nil, fmt.Errorf("invalid wildcard domain name: %s, it must have at least one character after '*.'", endpoint)
			}
		default:
			return nil, fmt.Errorf("invalid wildcard domain name: %s, it must have at most one wildcard character", endpoint)
		}
	}
	return endpoints, nil
}

// validateHost checks if the host is matching endpoints supported by the auth
// provider.
func validateHost(host string, endpoints []string) error {
	for _, endpoint := range endpoints {
		if endpoint[0] == '*' {
			if _, zone, ok := strings.Cut(host, "."); ok && zone == endpoint[2:] {
				return nil
			}
		}
		if host == endpoint {
			return nil
		}
	}
	return fmt.Errorf("the artifact host %s is not in the scope of the store auth provider", host)
}
