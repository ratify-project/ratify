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
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	provider "github.com/deislabs/ratify/pkg/referrerstore/oras/authprovider"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

type AwsIrsaBasicProviderFactory struct{}
type awsIrsaBasicAuthProvider struct {
	ecrAuthToken EcrAuthToken
	providerName string
}

type awsIrsaBasicAuthProviderConf struct {
	Name string `json:"name"`
}

const (
	awsIrsaEcrAuthProviderName string = "aws-irsa-basic"
)

// init calls Register for AWS IRSA Basic Auth provider
func init() {
	provider.Register(awsIrsaEcrAuthProviderName, &AwsIrsaBasicProviderFactory{})
}

func getEcrAuthToken() (EcrAuthToken, error) {
	region := os.Getenv("AWS_REGION")
	roleArn := os.Getenv("AWS_ROLE_ARN")
	tokenFilePath := os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE")

	if region == "" || roleArn == "" || tokenFilePath == "" {
		return EcrAuthToken{}, fmt.Errorf("required environment variables not set, AWS_REGION: %s, AWS_ROLE_ARN: %s, AWS_WEB_IDENTITY_TOKEN_FILE: %s", region, roleArn, tokenFilePath)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return EcrAuthToken{}, fmt.Errorf("failed to load default config: %v", err)
	}

	//client := sts.NewFromConfig(cfg)
	//
	//credsCache := aws.NewCredentialsCache(stscreds.NewWebIdentityRoleProvider(
	//	client,
	//	roleArn,
	//	stscreds.IdentityTokenFile(tokenFilePath),
	//	func(o *stscreds.WebIdentityRoleOptions) {
	//		o.RoleSessionName = awsSessionName
	//	}))
	//
	//creds, err := credsCache.Retrieve(context.TODO())
	//if err != nil {
	//	return aws.Credentials{}, fmt.Errorf("credentials not returned by AWS credential cache: %v", err)
	//}
	//
	//if !creds.HasKeys() || creds.SessionToken == "" {
	//	return aws.Credentials{}, fmt.Errorf("credential keys not returned")
	//}

	ecrClient := ecr.NewFromConfig(cfg)
	authOuput, err := ecrClient.GetAuthorizationToken(context.TODO(), nil)
	if err != nil {
		return EcrAuthToken{}, fmt.Errorf("could not reteive ECR auth token collection: %v", err)
	}

	//ecrAuthToken := tokenCollection.AuthorizationData[0]
	//rawDecodedToken, err := base64.StdEncoding.DecodeString(*ecrAuthToken.AuthorizationToken)
	//if err != nil {
	//	return awsIrsaBasicAuthProvider{}, fmt.Errorf("could not decode ECR auth token: %v", err)
	//}
	//
	//decodedAuthCreds := strings.Split(string(rawDecodedToken), ":")
	//
	//if decodedAuthCreds[0] == "" || decodedAuthCreds[1] == "" {
	//	return awsIrsaBasicAuthProvider{}, fmt.Errorf("empty ECR credentials returned")
	//}
	//
	//auth := awsIrsaBasicAuthProvider{}
	//auth.userName = decodedAuthCreds[0]
	//auth.password = decodedAuthCreds[1]
	//auth.proxyEndpoint = *ecrAuthToken.ProxyEndpoint
	//auth.providerName = awsIrsaAuthProviderName
	//auth.expiry = *ecrAuthToken.ExpiresAt

	return EcrAuthToken{AuthData: authOuput.AuthorizationData[0]}, nil
}

// Create returns an AwsIrsaBasicAuthProvider
func (s *AwsIrsaBasicProviderFactory) Create(authProviderConfig provider.AuthProviderConfig) (provider.AuthProvider, error) {
	conf := awsIrsaBasicAuthProviderConf{}
	authProviderConfigBytes, err := json.Marshal(authProviderConfig)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(authProviderConfigBytes, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse auth provider configuration: %v", err)
	}

	// Build auth provider from AWS IRSA ECR
	authData, err := getEcrAuthToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get ECR auth data: %v", err)
	}

	return &awsIrsaBasicAuthProvider{
		ecrAuthToken: authData,
		providerName: awsIrsaEcrAuthProviderName,
	}, nil
}

// Enabled checks for non empty AWS IAM creds
func (d *awsIrsaBasicAuthProvider) Enabled(ctx context.Context) bool {
	creds, err := d.ecrAuthToken.BasicAuthCreds()
	if creds == nil || err != nil {
		logrus.Errorf("error getting basic ECR creds: %v", err)
		return false
	}

	if len(creds) < 2 {
		return false
	}

	if creds[0] == "" || creds[1] == "" {
		return false
	}

	if d.providerName == "" {
		return false
	}

	return true
}

// Provide returns the credentials for a specified artifact.
// Uses AWS IRSA to retrieve creds from IRSA credential chain
func (d *awsIrsaBasicAuthProvider) Provide(ctx context.Context, artifact string) (provider.AuthConfig, error) {
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
		logrus.Info("successfully refreshed ECS auth token")
	}

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
