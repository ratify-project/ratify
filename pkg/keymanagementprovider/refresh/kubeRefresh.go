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

package refresh

import (
	"context"
	"fmt"
	"maps"
	"time"

	re "github.com/ratify-project/ratify/errors"
	kmp "github.com/ratify-project/ratify/pkg/keymanagementprovider"
	nv "github.com/ratify-project/ratify/pkg/verifier/notation"
	"github.com/sirupsen/logrus"
	ctrl "sigs.k8s.io/controller-runtime"
)

type KubeRefresher struct {
	Provider                kmp.KeyManagementProvider
	ProviderType            string
	ProviderRefreshInterval string
	Resource                string
	Result                  ctrl.Result
	Status                  kmp.KeyManagementProviderStatus
	CRLHandler              nv.RevocationFactory
}

// Register registers the kubeRefresher factory
func init() {
	Register(KubeRefresherType, &KubeRefresher{})
}

// Refresh the certificates/keys for the key management provider by calling the GetCertificates and GetKeys methods
func (kr *KubeRefresher) Refresh(ctx context.Context) error {
	logger := logrus.WithContext(ctx)

	// fetch certificates and store in map
	certificates, certAttributes, err := kr.Provider.GetCertificates(ctx)
	if err != nil {
		kmpErr := re.ErrorCodeKeyManagementProviderFailure.WithError(err).WithDetail(fmt.Sprintf("Unable to fetch certificates from key management provider [%s] of type [%s]", kr.Resource, kr.ProviderType))
		kmp.SetCertificateError(kr.Resource, err)
		return kmpErr
	}

	// fetch CRLs and cache them
	crlFetcher, err := kr.CRLHandler.NewFetcher()
	if err != nil {
		// log error and continue
		logger.Warnf("Unable to create CRL fetcher for key management provider %s of type %s with error: %v", kr.Resource, kr.ProviderType, err)
	}
	for _, cert := range certificates {
		nv.CacheCRL(ctx, cert, crlFetcher)
	}
	// fetch keys and store in map
	keys, keyAttributes, err := kr.Provider.GetKeys(ctx)
	if err != nil {
		kmpErr := re.ErrorCodeKeyManagementProviderFailure.WithError(err).WithDetail(fmt.Sprintf("Unable to fetch keys from key management provider [%s] of type [%s]", kr.Resource, kr.ProviderType))
		kmp.SetKeyError(kr.Resource, err)
		return kmpErr
	}

	kmp.SaveSecrets(kr.Resource, kr.ProviderType, keys, certificates)
	// merge certificates and keys status into one
	maps.Copy(keyAttributes, certAttributes)
	kr.Status = keyAttributes

	logger.Infof("%v certificate(s) & %v key(s) fetched for key management provider %v", len(certificates), len(keys), kr.Resource)

	// Resource is not refreshable, returning empty result and no error to indicate we’ve successfully reconciled this object
	// will not reconcile again unless resource is recreated
	if !kr.Provider.IsRefreshable() {
		return nil
	}

	// if interval is not set, disable refresh
	if kr.ProviderRefreshInterval == "" {
		return nil
	}

	// resource is refreshable, requeue after interval
	intervalDuration, err := time.ParseDuration(kr.ProviderRefreshInterval)
	if err != nil {
		kmpErr := re.ErrorCodeKeyManagementProviderFailure.WithError(err).WithDetail(fmt.Sprintf("Unable to parse interval duration for key management provider [%s] of type [%s]", kr.Resource, kr.ProviderType))
		return kmpErr
	}

	logger.Info("Reconciled KeyManagementProvider", "intervalDuration", intervalDuration)
	kr.Result = ctrl.Result{RequeueAfter: intervalDuration}

	return nil
}

// GetResult returns the result of the refresh as a ctrl.Result
func (kr *KubeRefresher) GetResult() interface{} {
	return kr.Result
}

func (kr *KubeRefresher) GetStatus() interface{} {
	return kr.Status
}

// Create creates a new KubeRefresher instance
func (kr *KubeRefresher) Create(config RefresherConfig) (Refresher, error) {
	return &KubeRefresher{
		Provider:                config.Provider,
		ProviderType:            config.ProviderType,
		ProviderRefreshInterval: config.ProviderRefreshInterval,
		Resource:                config.Resource,
		CRLHandler:              nv.CreateCRLHandlerFromConfig(),
	}, nil
}
