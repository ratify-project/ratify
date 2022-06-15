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
	"testing"
	"time"
)

const (
	testToken = "TEST_TOKEN"
)

// Verifies that userName, sessionToken, and
// environment variables are properly set
func TestAwsIrsaBasicAuthProvider_Enabled(t *testing.T) {
	authProvider := awsIrsaBasicAuthProvider{
		userName:     awsUserName,
		sessionToken: testToken,
		providerName: awsIrsaAuthProviderName,
		expiry:       time.Now(),
	}

	ctx := context.Background()

	if !authProvider.Enabled(ctx) {
		t.Fatal("enabled should have returned true but returned false")
	}

	authProvider.userName = ""
	if authProvider.Enabled(ctx) {
		t.Fatal("enabled should have returned false but returned true")
	}

	authProvider.userName = awsUserName
	authProvider.sessionToken = ""
	if authProvider.Enabled(ctx) {
		t.Fatal("enabled should have returned false but returned true")
	}

	authProvider.sessionToken = testToken
	authProvider.providerName = ""
	if authProvider.Enabled(ctx) {
		t.Fatal("enabled should have returned false but returned true")
	}
}
