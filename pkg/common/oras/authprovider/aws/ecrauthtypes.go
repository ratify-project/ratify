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

package aws

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
)

// EcrAuthToken provides helper functions for ECR auth token data
type EcrAuthToken struct {
	AuthData map[string]types.AuthorizationData
}

// Expiry returns the expiry time
func (e EcrAuthToken) Expiry(host string) time.Time {
	return *e.AuthData[host].ExpiresAt
}

// ProxyEndpoint returns the authdata proxy endpoint
func (e EcrAuthToken) ProxyEndpoint(host string) string {
	return *e.AuthData[host].ProxyEndpoint
}

// BasicAuthCreds returns a string array of the basic creds
func (e EcrAuthToken) BasicAuthCreds(host string) ([]string, error) {
	rawDecodedToken, err := base64.StdEncoding.DecodeString(*e.AuthData[host].AuthorizationToken)
	if err != nil {
		return nil, fmt.Errorf("could not decode ECR auth token: %w", err)
	}

	decodedAuthCreds := strings.Split(string(rawDecodedToken), ":")

	return decodedAuthCreds, nil
}
