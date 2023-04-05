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

package azureauth

import (
	"context"
	"fmt"
	"os"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

// Source: https://github.com/Azure/azure-workload-identity/blob/d126293e3c7c669378b225ad1b1f29cf6af4e56d/examples/msal-go/token_credential.go#L25
func GetAADAccessToken(ctx context.Context, tenantID, clientID, scope string) (confidential.AuthResult, error) {
	// Azure AD Workload Identity webhook will inject the following env vars:
	// 	AZURE_FEDERATED_TOKEN_FILE is the service account token path
	// 	AZURE_AUTHORITY_HOST is the AAD authority hostname

	/*tokenFilePath := os.Getenv("AZURE_FEDERATED_TOKEN_FILE")
	authority := os.Getenv("AZURE_AUTHORITY_HOST")

	if tokenFilePath == "" || authority == "" {
		return confidential.AuthResult{}, fmt.Errorf("required environment variables not set, AZURE_FEDERATED_TOKEN_FILE: %s, AZURE_AUTHORITY_HOST: %s", tokenFilePath, authority)
	}*/
	authority := "https://login.microsoftonline.com/"
	cred := confidential.NewCredFromAssertionCallback(func(context.Context, confidential.AssertionRequestOptions) (string, error) {
		// read the service account token from the filesystem
		return "eyJhbGciOiJSUzI1NiIsImtpZCI6Ii0yRS0tOW1jcDFFMG4wc1hxREhTLXk5Sjg3T0ZyaEtLQ1NXUUhZVlBPTzAifQ.eyJhdWQiOlsiYXBpOi8vQXp1cmVBRFRva2VuRXhjaGFuZ2UiXSwiZXhwIjoxNjgwNjk3ODQyLCJpYXQiOjE2ODA2OTQyNDIsImlzcyI6Imh0dHBzOi8vb2lkYy5wcm9kLWFrcy5henVyZS5jb20vYmVlMTUzMjgtMmZhYS00YWY5LTk5NmEtOTU2NDQ1N2EyZjA3LyIsImt1YmVybmV0ZXMuaW8iOnsibmFtZXNwYWNlIjoiZGVmYXVsdCIsInBvZCI6eyJuYW1lIjoicmF0aWZ5LTY0ZmY2OGJiZC14OXFjNCIsInVpZCI6ImU2YTI2ODlkLThjMzMtNGQyZi05MzM0LTIwOTJkMmI3NzM2NyJ9LCJzZXJ2aWNlYWNjb3VudCI6eyJuYW1lIjoid2xpZHNhIiwidWlkIjoiN2ZiMzg5YzItMGQyYS00MTBkLTg3YzgtODI0MWM3YzM5OTBhIn19LCJuYmYiOjE2ODA2OTQyNDIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OndsaWRzYSJ9.VcIzIB9R67HauvlSdAOc34ax4y2UFJbdzskhgkyK6y6nd0zSE5nTxl_4TxmUpYanuFzw31CanOhOsOSXttrEruhdFIv3lAk2lrX42AJ6FkjnKjEMmJ9uHZgqZiwSjZs2hb1l4OZHKAglPqqEpps3YPxGHmwyJWLLV6AYcppLZuGepgeUEjHlAd9Nsdoal3JaGW3n7n1aWHyjEVq_b57eaNV7ubTOwCsUtswQYVeQTW82QxokWLOdRCmOZ9n5JVwJaFwwU64cayZoheNDRDbDJPufl0GwtgXxaW94Q9Lv4hmOPacQgM9B7_MyAyn3T4sdCU2Bxic-fNjIQwNnN90Nh4P7-RDM0MRyKy_vwGpNQX_FrJaI_E9vIBRp9h7phgfbQn---ltvrFUQSTtJ5qHkvzaAJHAZEpwEYQqrsoMbATyjRDc49dT2pqA2yq0aLDtw5AvEpEBRTCqOX8MbWPVgpCY5E1GFG2ufKLQBGcjwzc9os8ZsFcgc8jEISV8cRWVRI9g-KxlnG253eJw6frNl6jhEDekznrKk1d_N5cHrJV-VIxBk-uB5P35t9dlTO0NJz_-oPF8PlCKMlmaJJPzQGuMCwcLa3K_xoNipZAWusxU3HC7XroRnB_r9v01ToaiTWUKmrmp7i_Qm-jKdTE_rUzf9OsJDSt9t5xg9bzavnhc", nil
	})

	// create the confidential client to request an AAD token
	confidentialClientApp, err := confidential.New(
		clientID,
		cred,
		confidential.WithAuthority(fmt.Sprintf("%s%s/oauth2/token", authority, tenantID)))
	if err != nil {
		return confidential.AuthResult{}, fmt.Errorf("failed to create confidential client app: %w", err)
	}

	result, err := confidentialClientApp.AcquireTokenByCredential(ctx, []string{scope})
	if err != nil {
		return confidential.AuthResult{}, fmt.Errorf("failed to acquire AAD token: %w", err)
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
