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
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"strings"
	"time"
)

// EcrAuthToken provides helper functions for ECR auth token data
type EcrAuthToken struct {
	AuthData types.AuthorizationData
}

// Expiry
func (e EcrAuthToken) Expiry() time.Time {
	return *e.AuthData.ExpiresAt
}

// ProxyEndpoint
func (e EcrAuthToken) ProxyEndpoint() string {
	return *e.AuthData.ProxyEndpoint
}

// BasicAuthCreds
func (e EcrAuthToken) BasicAuthCreds() ([]string, error) {
	rawDecodedToken, err := base64.StdEncoding.DecodeString(*e.AuthData.AuthorizationToken)
	if err != nil {
		return nil, fmt.Errorf("could not decode ECR auth token: %v", err)
	}

	decodedAuthCreds := strings.Split(string(rawDecodedToken), ":")

	return decodedAuthCreds, nil
}
