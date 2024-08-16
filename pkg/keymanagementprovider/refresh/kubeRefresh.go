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
	"encoding/json"
	"fmt"
	"maps"
	"time"

	configv1beta1 "github.com/ratify-project/ratify/api/v1beta1"
	"github.com/ratify-project/ratify/internal/constants"
	cutils "github.com/ratify-project/ratify/pkg/controllers/utils"
	kmp "github.com/ratify-project/ratify/pkg/keymanagementprovider"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KubeRefresher struct {
	client.Client
	Request ctrl.Request
	Result  ctrl.Result
}

// Register registers the kubeRefresher factory
func init() {
	Register(KubeRefresherType, &KubeRefresher{})
}

// Refresh the certificates/keys for the key management provider by calling the GetCertificates and GetKeys methods
func (kr *KubeRefresher) Refresh(ctx context.Context) error {
	logger := logrus.WithContext(ctx)

	var resource = kr.Request.Name
	var keyManagementProvider configv1beta1.KeyManagementProvider

	logger.Infof("reconciling cluster key management provider '%v'", resource)

	if err := kr.Get(ctx, kr.Request.NamespacedName, &keyManagementProvider); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Infof("deletion detected, removing key management provider %v", resource)
			kmp.DeleteCertificatesFromMap(resource)
			kmp.DeleteKeysFromMap(resource)
		} else {
			logger.Error(err, "unable to fetch key management provider")
		}

		kr.Result = ctrl.Result{}

		return client.IgnoreNotFound(err)
	}

	lastFetchedTime := metav1.Now()
	isFetchSuccessful := false

	// get certificate store list to check if certificate store is configured
	// TODO: remove check in v2.0.0+
	var certificateStoreList configv1beta1.CertificateStoreList
	if err := kr.List(ctx, &certificateStoreList); err != nil {
		logger.Error(err, "unable to list certificate stores")
		kr.Result = ctrl.Result{}
		return err
	}

	if len(certificateStoreList.Items) > 0 {
		// Note: for backwards compatibility in upgrade scenarios, Ratify will only log a warning statement.
		logger.Warn("Certificate Store already exists. Key management provider and certificate store should not be configured together. Please migrate to key management provider and delete certificate store.")
	}

	provider, err := cutils.SpecToKeyManagementProvider(keyManagementProvider.Spec.Parameters.Raw, keyManagementProvider.Spec.Type)
	if err != nil {
		writeKMProviderStatus(ctx, kr, &keyManagementProvider, logger, isFetchSuccessful, err.Error(), lastFetchedTime, nil)
		kr.Request = ctrl.Request{}
		return err
	}

	// fetch certificates and store in map
	certificates, certAttributes, err := provider.GetCertificates(ctx)
	if err != nil {
		writeKMProviderStatus(ctx, kr, &keyManagementProvider, logger, isFetchSuccessful, err.Error(), lastFetchedTime, nil)
		kr.Request = ctrl.Request{}
		return fmt.Errorf("error fetching certificates in KMProvider %v with %v provider, error: %w", resource, keyManagementProvider.Spec.Type, err)
	}

	// fetch keys and store in map
	keys, keyAttributes, err := provider.GetKeys(ctx)
	if err != nil {
		writeKMProviderStatus(ctx, kr, &keyManagementProvider, logger, isFetchSuccessful, err.Error(), lastFetchedTime, nil)
		kr.Request = ctrl.Request{}
		return fmt.Errorf("error fetching keys in KMProvider %v with %v provider, error: %w", resource, keyManagementProvider.Spec.Type, err)
	}
	kmp.SetCertificatesInMap(resource, certificates)
	kmp.SetKeysInMap(resource, keyManagementProvider.Spec.Type, keys)
	// merge certificates and keys status into one
	maps.Copy(keyAttributes, certAttributes)
	isFetchSuccessful = true
	emptyErrorString := ""
	writeKMProviderStatus(ctx, kr, &keyManagementProvider, logger, isFetchSuccessful, emptyErrorString, lastFetchedTime, keyAttributes)

	logger.Infof("%v certificate(s) & %v key(s) fetched for key management provider %v", len(certificates), len(keys), resource)

	// returning empty result and no error to indicate we’ve successfully reconciled this object
	// will not reconcile again unless resource is recreated
	if !provider.IsRefreshable() {
		kr.Request = ctrl.Request{}
		return nil
	}

	// if interval is not set, disable refresh
	if keyManagementProvider.Spec.RefreshInterval == "" {
		kr.Result = ctrl.Result{}
		return nil
	}
	// resource is refreshable, requeue after interval
	intervalDuration, err := time.ParseDuration(keyManagementProvider.Spec.RefreshInterval)
	if err != nil {
		logger.Error(err, "unable to parse interval duration")
		kr.Result = ctrl.Result{}
		return err
	}

	logger.Info("Reconciled KeyManagementProvider", "intervalDuration", intervalDuration)
	kr.Result = ctrl.Result{RequeueAfter: intervalDuration}

	return nil
}

// GetResult returns the result of the refresh as a ctrl.Result
func (kr *KubeRefresher) GetResult() interface{} {
	return kr.Result
}

// Create creates a new KubeRefresher instance
func (kr *KubeRefresher) Create(config map[string]interface{}) (Refresher, error) {
	client, ok := config["client"].(client.Client)
	if !ok {
		return nil, fmt.Errorf("client is required in config")
	}

	request, ok := config["request"].(ctrl.Request)
	if !ok {
		return nil, fmt.Errorf("request is required in config")
	}

	return &KubeRefresher{
		Client:  client,
		Request: request,
	}, nil
}

func writeKMProviderStatus(ctx context.Context, r client.StatusClient, keyManagementProvider *configv1beta1.KeyManagementProvider, logger *logrus.Entry, isSuccess bool, errorString string, operationTime metav1.Time, kmProviderStatus kmp.KeyManagementProviderStatus) {
	if isSuccess {
		updateKMProviderSuccessStatus(keyManagementProvider, &operationTime, kmProviderStatus)
	} else {
		updateKMProviderErrorStatus(keyManagementProvider, errorString, &operationTime)
	}
	if statusErr := r.Status().Update(ctx, keyManagementProvider); statusErr != nil {
		logger.Error(statusErr, ",unable to update key management provider error status")
	}
}

// updateKMProviderErrorStatus updates the key management provider status with error, brief error and last fetched time
func updateKMProviderErrorStatus(keyManagementProvider *configv1beta1.KeyManagementProvider, errorString string, operationTime *metav1.Time) {
	// truncate brief error string to maxBriefErrLength
	briefErr := errorString
	if len(errorString) > constants.MaxBriefErrLength {
		briefErr = fmt.Sprintf("%s...", errorString[:constants.MaxBriefErrLength])
	}
	keyManagementProvider.Status.IsSuccess = false
	keyManagementProvider.Status.Error = errorString
	keyManagementProvider.Status.BriefError = briefErr
	keyManagementProvider.Status.LastFetchedTime = operationTime
}

// updateKMProviderSuccessStatus updates the key management provider status if status argument is non nil
// Success status includes last fetched time and other provider-specific properties
func updateKMProviderSuccessStatus(keyManagementProvider *configv1beta1.KeyManagementProvider, lastOperationTime *metav1.Time, kmProviderStatus kmp.KeyManagementProviderStatus) {
	keyManagementProvider.Status.IsSuccess = true
	keyManagementProvider.Status.Error = ""
	keyManagementProvider.Status.BriefError = ""
	keyManagementProvider.Status.LastFetchedTime = lastOperationTime

	if kmProviderStatus != nil {
		jsonString, _ := json.Marshal(kmProviderStatus)

		raw := runtime.RawExtension{
			Raw: jsonString,
		}
		keyManagementProvider.Status.Properties = raw
	}
}
