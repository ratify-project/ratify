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
// Source: https://github.com/Azure/secrets-store-csi-driver-provider-azure/tree/release-1.4/pkg/auth
import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/deislabs/ratify/pkg/utils/azureauth"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
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

func getAuthorizerForWorkloadIdentity(ctx context.Context, tenantID, clientID, resource string) (autorest.Authorizer, error) {
	scope := resource
	// .default needs to be added to the scope
	if !strings.Contains(resource, ".default") {
		scope = fmt.Sprintf("%s/.default", resource)
	}

	result, err := azureauth.GetAADAccessToken(ctx, tenantID, clientID, scope)
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

// Vendored from https://github.com/Azure/go-autorest/blob/79575dd7ba2e88e7ce7ab84e167ec6653dcb70c1/autorest/adal/token.go
// converts expires_on to the number of seconds
func parseExpiresOn(s interface{}) (json.Number, error) {
	// the JSON unmarshaler treats JSON numbers unmarshaled into an interface{} as float64
	asFloat64, ok := s.(float64)
	if ok {
		// this is the number of seconds as int case
		return json.Number(strconv.FormatInt(int64(asFloat64), 10)), nil
	}
	asStr, ok := s.(string)
	if !ok {
		return "", fmt.Errorf("unexpected expires_on type %T", s)
	}
	// convert the expiration date to the number of seconds from the unix epoch
	timeToDuration := func(t time.Time) json.Number {
		return json.Number(strconv.FormatInt(t.UTC().Unix(), 10))
	}
	if _, err := json.Number(asStr).Int64(); err == nil {
		// this is the number of seconds case, no conversion required
		return json.Number(asStr), nil
	} else if eo, err := time.Parse(expiresOnDateFormatPM, asStr); err == nil {
		return timeToDuration(eo), nil
	} else if eo, err := time.Parse(expiresOnDateFormat, asStr); err == nil {
		return timeToDuration(eo), nil
	}
	return json.Number(""), fmt.Errorf("unknown expires_on format %s", asStr)
}
