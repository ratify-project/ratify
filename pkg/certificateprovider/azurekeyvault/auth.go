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

package azurekeyvault

// This class is based on implementation from  azure secret store csi provider
// Source: https://github.com/Azure/secrets-store-csi-driver-provider-azure/blob/master/pkg/provider/
import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
	"github.com/pkg/errors"
)

const (
	// the format for expires_on in UTC with AM/PM
	expiresOnDateFormatPM = "1/2/2006 15:04:05 PM +00:00"
	// the format for expires_on in UTC without AM/PM
	expiresOnDateFormat = "1/2/2006 15:04:05 +00:00"

	tokenTypeBearer = "Bearer"
	// For Azure AD Workload Identity, the audience recommended for use is
	// "api://AzureADTokenExchange"
	DefaultTokenAudience = "api://AzureADTokenExchange" //nolint
)

// authResult contains the subset of results from token acquisition operation in ConfidentialClientApplication
// For details see https://aka.ms/msal-net-authenticationresult
type authResult struct {
	accessToken    string
	expiresOn      time.Time
	grantedScopes  []string
	declinedScopes []string
}

func getAuthorizerForWorkloadIdentity(ctx context.Context, tenantID, clientID, resource, aadEndpoin string) (autorest.Authorizer, error) {

	scope := resource
	// .default needs to be added to the scope
	if !strings.Contains(resource, ".default") {
		scope = fmt.Sprintf("%s/.default", resource)
	}
	result, err := getAADAccessToken(ctx, tenantID, clientID, scope)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire token: %w", err)
	}

	token := adal.Token{
		AccessToken: result.AccessToken,
		Resource:    resource,
		Type:        tokenTypeBearer,
	}
	token.ExpiresOn, err = parseExpiresOn(result.ExpiresOn.UTC().Local().Format(expiresOnDateFormat))
	if err != nil {
		return nil, fmt.Errorf("failed to parse expires_on: %w", err)
	}

	return autorest.NewBearerAuthorizer(authResult{
		accessToken:    result.AccessToken,
		expiresOn:      result.ExpiresOn,
		grantedScopes:  result.GrantedScopes,
		declinedScopes: result.DeclinedScopes,
	}), nil
}

// OAuthToken implements the OAuthTokenProvider interface.  It returns the current access token.
func (ar authResult) OAuthToken() string {
	return ar.accessToken
}

// Vendored from https://github.com/Azure/go-autorest/blob/def88ef859fb980eff240c755a70597bc9b490d0/autorest/adal/token.go
// converts expires_on to the number of seconds
func parseExpiresOn(s string) (json.Number, error) {
	// convert the expiration date to the number of seconds from now
	timeToDuration := func(t time.Time) json.Number {
		dur := t.Sub(time.Now().UTC())
		return json.Number(strconv.FormatInt(int64(dur.Round(time.Second).Seconds()), 10))
	}
	if _, err := strconv.ParseInt(s, 10, 64); err == nil {
		// this is the number of seconds case, no conversion required
		return json.Number(s), nil
	} else if eo, err := time.Parse(expiresOnDateFormatPM, s); err == nil {
		return timeToDuration(eo), nil
	} else if eo, err := time.Parse(expiresOnDateFormat, s); err == nil {
		return timeToDuration(eo), nil
	} else {
		// unknown format
		return json.Number(""), err
	}
}

// Source: https://github.com/Azure/azure-workload-identity/blob/d126293e3c7c669378b225ad1b1f29cf6af4e56d/examples/msal-go/token_credential.go#L25
func getAADAccessToken(ctx context.Context, tenantID string, clientID string, scope string) (confidential.AuthResult, error) {
	// Azure AD Workload Identity webhook will inject the following env vars:
	// 	AZURE_CLIENT_ID with the clientID set in the service account annotation
	// 	AZURE_TENANT_ID with the tenantID set in the service account annotation. If not defined, then
	// 	the tenantID provided via azure-wi-webhook-config for the webhook will be used.
	// 	AZURE_FEDERATED_TOKEN_FILE is the service account token path
	// 	AZURE_AUTHORITY_HOST is the AAD authority hostname

	tokenFilePath := os.Getenv("AZURE_FEDERATED_TOKEN_FILE")
	authority := os.Getenv("AZURE_AUTHORITY_HOST")

	if clientID == "" || tokenFilePath == "" || authority == "" {
		return confidential.AuthResult{}, fmt.Errorf("required environment variables not set, AZURE_CLIENT_ID: %s, AZURE_FEDERATED_TOKEN_FILE: %s, AZURE_AUTHORITY_HOST: %s", clientID, tokenFilePath, authority)
	}

	// read the service account token from the filesystem
	signedAssertion, err := readJWTFromFS(tokenFilePath)
	if err != nil {
		return confidential.AuthResult{}, errors.Wrap(err, "failed to read service account token")
	}

	cred, err := confidential.NewCredFromAssertion(signedAssertion)
	if err != nil {
		return confidential.AuthResult{}, errors.Wrap(err, "failed to create confidential creds")
	}

	// create the confidential client to request an AAD token
	confidentialClientApp, err := confidential.New(
		clientID,
		cred,
		confidential.WithAuthority(fmt.Sprintf("%s%s/oauth2/token", authority, tenantID)))
	if err != nil {
		return confidential.AuthResult{}, errors.Wrap(err, "failed to create confidential client app")
	}

	result, err := confidentialClientApp.AcquireTokenByCredential(ctx, []string{scope})
	if err != nil {
		return confidential.AuthResult{}, errors.Wrap(err, "failed to acquire AAD token")
	}

	return result, nil
}

// readJWTFromFS reads the jwt from file system
// Source: https://github.com/Azure/azure-workload-identity/blob/d126293e3c7c669378b225ad1b1f29cf6af4e56d/examples/msal-go/token_credential.go#L88
func readJWTFromFS(tokenFilePath string) (string, error) {
	token, err := os.ReadFile(tokenFilePath)
	if err != nil {
		return "", err
	}
	return string(token), nil
}
