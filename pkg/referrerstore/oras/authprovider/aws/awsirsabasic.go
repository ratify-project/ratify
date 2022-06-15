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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	provider "github.com/deislabs/ratify/pkg/referrerstore/oras/authprovider"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

type AwsIrsaBasicProviderFactory struct{}
type awsIrsaBasicAuthProvider struct {
	userName     string
	sessionToken string
	providerName string
	expiry       time.Time
}

type awsIrsaBasicAuthProviderConf struct {
	Name string `json:"name"`
}

const (
	awsIrsaAuthProviderName string = "aws-irsa-basic"
	awsUserName             string = "AWS"
	awsSessionName          string = "ratify-irsa"
)

// init calls Register for AWS IRSA Basic Auth provider
func init() {
	provider.Register(awsIrsaAuthProviderName, &AwsIrsaBasicProviderFactory{})
}

func getIrsaCreds() (aws.Credentials, error) {
	region := os.Getenv("AWS_REGION")
	roleArn := os.Getenv("AWS_ROLE_ARN")
	tokenFilePath := os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE")

	if region == "" || roleArn == "" || tokenFilePath == "" {
		return aws.Credentials{}, fmt.Errorf("required environment variables not set, AWS_REGION: %s, AWS_ROLE_ARN: %s, AWS_WEB_IDENTITY_TOKEN_FILE: %s", region, roleArn, tokenFilePath)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		panic("failed to load config, " + err.Error())
	}

	client := sts.NewFromConfig(cfg)

	credsCache := aws.NewCredentialsCache(stscreds.NewWebIdentityRoleProvider(
		client,
		roleArn,
		stscreds.IdentityTokenFile(tokenFilePath),
		func(o *stscreds.WebIdentityRoleOptions) {
			o.RoleSessionName = awsSessionName
		}))

	creds, err := credsCache.Retrieve(context.TODO())
	if err != nil {
		return aws.Credentials{}, fmt.Errorf("credentials not returned by AWS credential cache: %v", err)
	}

	if !creds.HasKeys() || creds.SessionToken == "" {
		return aws.Credentials{}, fmt.Errorf("credential keys not returned")
	}

	return creds, nil
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

	// Get AWS creds
	creds, err := getIrsaCreds()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve usable creds: %v", err)
	}

	return &awsIrsaBasicAuthProvider{
		userName:     awsUserName,
		sessionToken: creds.SessionToken,
		providerName: awsIrsaAuthProviderName,
		expiry:       creds.Expires,
	}, nil
}

// Enabled checks for non empty AWS IAM creds
func (d *awsIrsaBasicAuthProvider) Enabled(ctx context.Context) bool {
	if d.userName == "" {
		return false
	}

	if d.sessionToken == "" {
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

	// need to refresh AWS IRSA session token if it's expired 5 mins from now
	t := time.Now().Add(time.Minute * 5)
	if t.After(d.expiry) {
		creds, err := getIrsaCreds()
		if err != nil {
			return provider.AuthConfig{}, errors.Wrap(err, "could not refresh session token")
		}
		d.sessionToken = creds.SessionToken
		d.expiry = creds.Expires
		logrus.Info("successfully refreshed AWS IRSA session token")
	}

	authConfig := provider.AuthConfig{
		Username:  awsUserName,
		Password:  d.sessionToken,
		Provider:  d,
		ExpiresOn: d.expiry,
	}

	return authConfig, nil
}
