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
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/containers/azcontainerregistry"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
	"github.com/ratify-project/ratify/internal/logger"
	"github.com/ratify-project/ratify/pkg/utils/azureauth"
)

func DefaultAuthClientFactory(serverURL string, options *azcontainerregistry.AuthenticationClientOptions) (AuthClient, error) {
	client, err := azcontainerregistry.NewAuthenticationClient(serverURL, options)
	if err != nil {
		return nil, err
	}
	return &AuthenticationClientWrapper{client: client}, nil
}

func DefaultGetAADAccessToken(ctx context.Context, tenantID, clientID, resource string) (confidential.AuthResult, error) {
	return azureauth.GetAADAccessToken(ctx, tenantID, clientID, resource)
}

func DefaultReportMetrics(ctx context.Context, duration int64, artifactHostName string) {
	logger.GetLogger(ctx, logOpt).Infof("Metrics Report: Duration=%dms, Host=%s", duration, artifactHostName)
}

type AuthenticationClientWrapper struct {
	client *azcontainerregistry.AuthenticationClient
}

func (w *AuthenticationClientWrapper) ExchangeAADAccessTokenForACRRefreshToken(ctx context.Context, grantType, service string, options *azcontainerregistry.AuthenticationClientExchangeAADAccessTokenForACRRefreshTokenOptions) (azcontainerregistry.AuthenticationClientExchangeAADAccessTokenForACRRefreshTokenResponse, error) {
	return w.client.ExchangeAADAccessTokenForACRRefreshToken(ctx, azcontainerregistry.PostContentSchemaGrantType(grantType), service, options)
}

type AuthClient interface {
	ExchangeAADAccessTokenForACRRefreshToken(ctx context.Context, grantType, service string, options *azcontainerregistry.AuthenticationClientExchangeAADAccessTokenForACRRefreshTokenOptions) (azcontainerregistry.AuthenticationClientExchangeAADAccessTokenForACRRefreshTokenResponse, error)
}
