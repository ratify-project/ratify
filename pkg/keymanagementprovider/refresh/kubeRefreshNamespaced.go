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
	re "github.com/ratify-project/ratify/errors"
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

type KubeRefresherNamespaced struct {
	client.Client
	Request ctrl.Request
	Result  ctrl.Result
}

// Register registers the kubeRefresherNamespaced factory
func init() {
	Register(KubeRefresherNamespacedType, &KubeRefresherNamespaced{})
}

// Refresh the certificates/keys for the key management provider by calling the GetCertificates and GetKeys methods
func (kr *KubeRefresherNamespaced) Refresh(ctx context.Context) error {
	logger := logrus.WithContext(ctx)

	var resource = kr.Request.NamespacedName.String()
	var keyManagementProvider configv1beta1.NamespacedKeyManagementProvider

	logger.Infof("reconciling namespaced key management provider '%v'", resource)

	if err := kr.Get(ctx, kr.Request.NamespacedName, &keyManagementProvider); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Infof("deletion detected, removing key management provider %v", resource)
			kmp.DeleteResourceFromMap(resource)
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
	// if certificate store is configured, return error. Only one of certificate store and key management provider can be configured
	if len(certificateStoreList.Items) > 0 {
		// Note: for backwards compatibility in upgrade scenarios, Ratify will only log a warning statement.
		logger.Warn("Certificate Store already exists. Key management provider and certificate store should not be configured together. Please migrate to key management provider and delete certificate store.")
	}

	provider, err := cutils.SpecToKeyManagementProvider(keyManagementProvider.Spec.Parameters.Raw, keyManagementProvider.Spec.Type)
	if err != nil {
		kmpErr := re.ErrorCodePluginInitFailure.WithError(err).WithDetail("Failed to create key management provider from CR")

		kmp.SetCertificateError(resource, kmpErr)
		kmp.SetKeyError(resource, kmpErr)
		writeKMProviderStatusNamespaced(ctx, kr, &keyManagementProvider, logger, isFetchSuccessful, &kmpErr, lastFetchedTime, nil)
		kr.Result = ctrl.Result{}
		return kmpErr
	}

	// fetch certificates and store in map
	certificates, certAttributes, err := provider.GetCertificates(ctx)
	if err != nil {
		kmpErr := re.ErrorCodeKeyManagementProviderFailure.WithError(err).WithDetail(fmt.Sprintf("Unable to fetch certificates from key management provider: %s of type: %s", resource, keyManagementProvider.Spec.Type))

		kmp.SetCertificateError(resource, kmpErr)
		writeKMProviderStatusNamespaced(ctx, kr, &keyManagementProvider, logger, isFetchSuccessful, &kmpErr, lastFetchedTime, nil)
		kr.Result = ctrl.Result{}
		return kmpErr
	}

	// fetch keys and store in map
	keys, keyAttributes, err := provider.GetKeys(ctx)
	if err != nil {
		kmpErr := re.ErrorCodeKeyManagementProviderFailure.WithError(err).WithDetail(fmt.Sprintf("Unable to fetch keys from key management provider: %s of type: %s", resource, keyManagementProvider.Spec.Type))

		kmp.SetKeyError(resource, kmpErr)
		writeKMProviderStatusNamespaced(ctx, kr, &keyManagementProvider, logger, isFetchSuccessful, &kmpErr, lastFetchedTime, nil)
		kr.Result = ctrl.Result{}
		return kmpErr
	}
	kmp.SaveSecrets(resource, keyManagementProvider.Spec.Type, keys, certificates)
	// merge certificates and keys status into one
	maps.Copy(keyAttributes, certAttributes)
	isFetchSuccessful = true
	writeKMProviderStatusNamespaced(ctx, kr, &keyManagementProvider, logger, isFetchSuccessful, nil, lastFetchedTime, keyAttributes)

	logger.Infof("%v certificate(s) & %v key(s) fetched for key management provider %v", len(certificates), len(keys), resource)

	if !provider.IsRefreshable() {
		kr.Result = ctrl.Result{}
		return nil
	}

	// if interval is not set, disable refresh
	if keyManagementProvider.Spec.RefreshInterval == "" {
		kr.Result = ctrl.Result{}
		return nil
	}

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

// GetResult returns the result of the refresh as ctrl.Result
func (kr *KubeRefresherNamespaced) GetResult() interface{} {
	return kr.Result
}

// Create creates a new instance of KubeRefresherNamespaced
func (kr *KubeRefresherNamespaced) Create(config map[string]interface{}) (Refresher, error) {
	client, ok := config["client"].(client.Client)
	if !ok {
		return nil, fmt.Errorf("client is required in config")
	}

	request, ok := config["request"].(ctrl.Request)
	if !ok {
		return nil, fmt.Errorf("request is required in config")
	}

	return &KubeRefresherNamespaced{
		Client:  client,
		Request: request,
	}, nil
}

// writeKMProviderStatus updates the status of the key management provider resource
func writeKMProviderStatusNamespaced(ctx context.Context, r client.StatusClient, keyManagementProvider *configv1beta1.NamespacedKeyManagementProvider, logger *logrus.Entry, isSuccess bool, err *re.Error, operationTime metav1.Time, kmProviderStatus kmp.KeyManagementProviderStatus) {
	if isSuccess {
		updateKMProviderSuccessStatusNamespaced(keyManagementProvider, &operationTime, kmProviderStatus)
	} else {
		updateKMProviderErrorStatusNamespaced(keyManagementProvider, err, &operationTime)
	}
	if statusErr := r.Status().Update(ctx, keyManagementProvider); statusErr != nil {
		logger.Error(statusErr, ",unable to update key management provider error status")
	}
}

// updateKMProviderErrorStatus updates the key management provider status with error, brief error and last fetched time
func updateKMProviderErrorStatusNamespaced(keyManagementProvider *configv1beta1.NamespacedKeyManagementProvider, err *re.Error, operationTime *metav1.Time) {
	keyManagementProvider.Status.IsSuccess = false
	keyManagementProvider.Status.Error = err.Error()
	keyManagementProvider.Status.BriefError = err.GetConciseError(constants.MaxBriefErrLength)
	keyManagementProvider.Status.LastFetchedTime = operationTime
}

// updateKMProviderSuccessStatus updates the key management provider status if status argument is non nil
// Success status includes last fetched time and other provider-specific properties
func updateKMProviderSuccessStatusNamespaced(keyManagementProvider *configv1beta1.NamespacedKeyManagementProvider, lastOperationTime *metav1.Time, kmProviderStatus kmp.KeyManagementProviderStatus) {
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
