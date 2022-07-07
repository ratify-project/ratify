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
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	provider "github.com/deislabs/ratify/pkg/referrerstore/oras/authprovider"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

type AwsEcrBasicProviderFactory struct{}
type awsEcrBasicAuthProvider struct {
	ecrAuthToken EcrAuthToken
	providerName string
}

type awsEcrBasicAuthProviderConf struct {
	Name string `json:"name"`
}

const (
	awsEcrAuthProviderName string = "awsEcrBasic"
	awsSessionName         string = "ratifyEcrBasicAuth"
)

// init calls Register for AWS IRSA Basic Auth provider
func init() {
	provider.Register(awsEcrAuthProviderName, &AwsEcrBasicProviderFactory{})
}

// Get ECR auth token from IRSA config
func getEcrAuthToken() (EcrAuthToken, error) {
	region := os.Getenv("AWS_REGION")
	roleArn := os.Getenv("AWS_ROLE_ARN")
	tokenFilePath := os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE")

	// Verify IRSA ENV is present
	if region == "" || roleArn == "" || tokenFilePath == "" {
		return EcrAuthToken{}, fmt.Errorf("required environment variables not set, AWS_REGION: %s, AWS_ROLE_ARN: %s, AWS_WEB_IDENTITY_TOKEN_FILE: %s", region, roleArn, tokenFilePath)
	}

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region),
		config.WithWebIdentityRoleCredentialOptions(func(options *stscreds.WebIdentityRoleOptions) {
			options.RoleSessionName = awsSessionName
		}))
	if err != nil {
		return EcrAuthToken{}, fmt.Errorf("failed to load default config: %v", err)
	}

	ecrClient := ecr.NewFromConfig(cfg)
	authOutput, err := ecrClient.GetAuthorizationToken(ctx, nil)
	if err != nil {
		return EcrAuthToken{}, fmt.Errorf("could not reteive ECR auth token collection: %v", err)
	}

	return EcrAuthToken{AuthData: authOutput.AuthorizationData[0]}, nil
}

// Create returns an AwsEcrBasicProvider
func (s *AwsEcrBasicProviderFactory) Create(authProviderConfig provider.AuthProviderConfig) (provider.AuthProvider, error) {
	conf := awsEcrBasicAuthProviderConf{}
	authProviderConfigBytes, err := json.Marshal(authProviderConfig)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(authProviderConfigBytes, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse auth provider configuration: %v", err)
	}

	// Build auth provider from AWS IRSA and ECR auth token
	authData, err := getEcrAuthToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get ECR auth data: %v", err)
	}

	return &awsEcrBasicAuthProvider{
		ecrAuthToken: authData,
		providerName: awsEcrAuthProviderName,
	}, nil
}

// Enabled checks for non empty AWS IAM creds
func (d *awsEcrBasicAuthProvider) Enabled(ctx context.Context) bool {
	creds, err := d.ecrAuthToken.BasicAuthCreds()
	if creds == nil || err != nil {
		logrus.Errorf("error getting basic ECR creds: %v", err)
		return false
	}

	if len(creds) < 2 {
		logrus.Error("creds array had incorrect length")
		return false
	}

	if creds[0] == "" || creds[1] == "" {
		logrus.Error("creds were empty")
		return false
	}

	if d.providerName == "" {
		logrus.Error("providerName was empty")
		return false
	}

	if d.ecrAuthToken.AuthData.ExpiresAt == nil {
		logrus.Error("expiry was nil")
		return false
	}

	return true
}

// Provide returns the credentials for a specified artifact.
// Uses AWS IRSA to retrieve creds from IRSA credential chain
func (d *awsEcrBasicAuthProvider) Provide(ctx context.Context, artifact string) (provider.AuthConfig, error) {
	if !d.Enabled(ctx) {
		return provider.AuthConfig{}, fmt.Errorf("AWS IRSA basic auth provider is not properly enabled")
	}

	// need to refresh AWS ECR credentials
	t := time.Now().Add(time.Minute * 5)
	if t.After(d.ecrAuthToken.Expiry()) || time.Now().After(d.ecrAuthToken.Expiry()) {
		authToken, err := getEcrAuthToken()
		if err != nil {
			return provider.AuthConfig{}, errors.Wrap(err, "could not refresh ECR auth token")
		}
		d.ecrAuthToken = authToken
		logrus.Info("successfully refreshed ECR auth token")
	}

	// Get ECR basic auth creds from auth data token
	creds, err := d.ecrAuthToken.BasicAuthCreds()
	if err != nil {
		return provider.AuthConfig{}, errors.Wrap(err, "could not retrieve ECR credentials")
	}

	authConfig := provider.AuthConfig{
		Username:  creds[0],
		Password:  creds[1],
		Provider:  d,
		ExpiresOn: d.ecrAuthToken.Expiry(),
	}

	return authConfig, nil
}
