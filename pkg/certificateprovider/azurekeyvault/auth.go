package azurekeyvault

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
	// Pod Identity podNameHeader
	podNameHeader = "podname"
	// Pod Identity podNamespaceHeader
	podNamespaceHeader = "podns"

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

	//question, what is token for?

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
	/*clientID := os.Getenv("AZURE_CLIENT_ID")
	tokenFilePath := os.Getenv("AZURE_FEDERATED_TOKEN_FILE")
	authority := os.Getenv("AZURE_AUTHORITY_HOST")

	if clientID == "" || tokenFilePath == "" || authority == "" {
		return confidential.AuthResult{}, fmt.Errorf("required environment variables not set, AZURE_CLIENT_ID: %s, AZURE_FEDERATED_TOKEN_FILE: %s, AZURE_AUTHORITY_HOST: %s", clientID, tokenFilePath, authority)
	}

	// read the service account token from the filesystem
	/*signedAssertion, err := readJWTFromFS(tokenFilePath)
	if err != nil {
		return confidential.AuthResult{}, errors.Wrap(err, "failed to read service account token")
	}*/
	clientID = "1c7ac023-5bf6-4916-83f2-96dd203e35a3"
	signedAssertion := "eyJhbGciOiJSUzI1NiIsImtpZCI6ImVoeVRBY1RYYkk4TjhfUmtPcjF3RmItRDRqcjYzbDBPXzRjb29YLWVwbXcifQ.eyJhdWQiOlsiYXBpOi8vQXp1cmVBRFRva2VuRXhjaGFuZ2UiXSwiZXhwIjoxNjcyOTQ0ODIxLCJpYXQiOjE2NzI5NDEyMjEsImlzcyI6Imh0dHBzOi8vb2lkYy5wcm9kLWFrcy5henVyZS5jb20vYmVlMTUzMjgtMmZhYS00YWY5LTk5NmEtOTU2NDQ1N2EyZjA3LyIsImt1YmVybmV0ZXMuaW8iOnsibmFtZXNwYWNlIjoiZGVmYXVsdCIsInBvZCI6eyJuYW1lIjoicmF0aWZ5LWM1NjRkNmRmNS1iZGxjdyIsInVpZCI6Ijk3NjY2MTUyLWQ5YjMtNGZhZS1iMzEzLTgzMzdjM2MwNjcwZSJ9LCJzZXJ2aWNlYWNjb3VudCI6eyJuYW1lIjoid2xpZHNhIiwidWlkIjoiODIxMmQ1NjQtNjkyNy00MWJiLTllY2MtNTBlM2IzNzBhOWFhIn19LCJuYmYiOjE2NzI5NDEyMjEsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OndsaWRzYSJ9.jQ3T1b_wBtDrYm0q91ES-5ggJIGwiXk0S8fzG-llTQaipsqNrxfOrgxHdei-EWoz0LkV0Y7iI3Qed6IcpjSYiE0Mognila-n6E5KKhCvvQixwSLMQzE-94syRj6nwBwhDXYAkV53wuYZBKVt2GhQYFPw--EhBw7dpeH1N6Il9t6hFauqqsX-swhOaHqDGiZ3FoU7Y9D9bPxGSmchty7ZH58Z9j1gFDozJAKcQyCB_u5EahFBVSuu56yeC_hBVnBZZvSfViRAcDyPjK7t1V50yLlSqR7xPTKQH_YUpznKQKcb_57Xe_SxseYjhSaifSvWvaLF8LJ4pqv3rjSvaa970ENEZ2YvKRvj2Afd-OlpW2WmBjmP9kOE0MP27qZ_j8B4DDj4mwD0NzvrfeQ_-kezGpslivWt4VOFrXyzhfYofAyiPTAOznYooGqa7eiZEqXRQDbiHAG50kZEh5QdYJbuaRHWMNdCOp9zOk8P74VCbgZe3l6HD406gHmepL095lIm0QI8MMllDVT0Rc1D7oDt8pA_hhBTkDBAJv0dCLGXYlrgLqqlMyxlrV-YG3RBQEEOQmTSJuHoBanz4ZIYyANQai_j3woQJtvP1k5skTnWU1qfJFfailEgNys_URhLUNHv589HL-TBOpzaK1jadefOIBCPbpqWS220kbK9whvz9jY"
	authority := "https://login.microsoftonline.com/"
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
