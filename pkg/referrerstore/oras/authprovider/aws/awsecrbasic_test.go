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
	"context"
	"encoding/base64"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"strings"
	"testing"
	"time"
)

const (
	testUsername = "AWS"
	testPassword = "eyJwYXlsb2FkIjoiOThPNTFqemhaUmZWVG"
	testProxy    = "PROXY_ENDPOINT"
)

// Verifies that awsEcrBasicAuthProvider is properly constructed by mocking EcrAuthToken

func mockAuthData() types.AuthorizationData {
	// Setup
	creds := []string{testUsername, testPassword}
	encoded := base64.StdEncoding.EncodeToString([]byte(strings.Join(creds, ":")))

	return types.AuthorizationData{
		AuthorizationToken: aws.String(encoded),
		ExpiresAt:          aws.Time(time.Now().Add(time.Minute * 10)),
		ProxyEndpoint:      aws.String(testProxy),
	}
}

func TestAwsEcrBasicAuthProvider_Enabled(t *testing.T) {

	authProvider := awsEcrBasicAuthProvider{
		ecrAuthToken: EcrAuthToken{mockAuthData()},
		providerName: awsEcrAuthProviderName,
	}

	ctx := context.Background()

	if !authProvider.Enabled(ctx) {
		t.Fatal("enabled should have returned true but returned false")
	}

	authProvider.providerName = ""
	if authProvider.Enabled(ctx) {
		t.Fatal("enabled should have returned false but returned true")
	}

	authProvider = awsEcrBasicAuthProvider{
		ecrAuthToken: EcrAuthToken{mockAuthData()},
		providerName: awsEcrAuthProviderName,
	}

	authProvider.ecrAuthToken.AuthData.AuthorizationToken = aws.String("")
	if authProvider.Enabled(ctx) {
		t.Fatal("enabled should have returned false but returned true")
	}

	authProvider = awsEcrBasicAuthProvider{
		ecrAuthToken: EcrAuthToken{mockAuthData()},
		providerName: awsEcrAuthProviderName,
	}

	authProvider.ecrAuthToken.AuthData.ExpiresAt = nil
	if authProvider.Enabled(ctx) {
		t.Fatal("enabled should have returned false but returned true")
	}
}
