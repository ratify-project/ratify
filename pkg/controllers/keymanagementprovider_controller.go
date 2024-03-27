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

package controllers

import (
	"context"
	"encoding/json"
	"fmt"

	_ "github.com/deislabs/ratify/pkg/keymanagementprovider/azurekeyvault" // register azure key vault key management provider
	_ "github.com/deislabs/ratify/pkg/keymanagementprovider/inline"        // register inline key management provider
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	configv1beta1 "github.com/deislabs/ratify/api/v1beta1"
	c "github.com/deislabs/ratify/config"
	re "github.com/deislabs/ratify/errors"
	"github.com/deislabs/ratify/pkg/keymanagementprovider"
	"github.com/deislabs/ratify/pkg/keymanagementprovider/config"
	"github.com/deislabs/ratify/pkg/keymanagementprovider/factory"
	"github.com/deislabs/ratify/pkg/keymanagementprovider/types"
	"github.com/sirupsen/logrus"
)

// KeyManagementProviderReconciler reconciles a KeyManagementProvider object
type KeyManagementProviderReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=keymanagementproviders,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=keymanagementproviders/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=keymanagementproviders/finalizers,verbs=update
func (r *KeyManagementProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logrus.WithContext(ctx)

	var resource = req.NamespacedName.String()
	var keyManagementProvider configv1beta1.KeyManagementProvider

	logger.Infof("reconciling key management provider '%v'", resource)

	if err := r.Get(ctx, req.NamespacedName, &keyManagementProvider); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Infof("deletion detected, removing key management provider %v", resource)
			keymanagementprovider.DeleteCertificatesFromMap(resource)
		} else {
			logger.Error(err, "unable to fetch key management provider")
		}

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	lastFetchedTime := metav1.Now()
	isFetchSuccessful := false

	// get certificate store list to check if certificate store is configured
	// TODO: remove check in v2.0.0+
	var certificateStoreList configv1beta1.CertificateStoreList
	if err := r.List(ctx, &certificateStoreList); err != nil {
		logger.Error(err, "unable to list certificate stores")
		return ctrl.Result{}, err
	}
	// if certificate store is configured, return error. Only one of certificate store and key management provider can be configured
	if len(certificateStoreList.Items) > 0 {
		err := re.ErrorCodeKeyManagementConflict.WithComponentType(re.KeyManagementProvider).WithPluginName(resource).WithDetail("certificate store already exists")
		// Note: for backwards compatibility in upgrade scenarios, Ratify will only log an error.
		logger.Error(err)
	}

	provider, err := specToKeyManagementProvider(keyManagementProvider.Spec)
	if err != nil {
		writeKMProviderStatus(ctx, r, &keyManagementProvider, logger, isFetchSuccessful, err.Error(), lastFetchedTime, nil)
		return ctrl.Result{}, err
	}

	certificates, certAttributes, err := provider.GetCertificates(ctx)
	if err != nil {
		writeKMProviderStatus(ctx, r, &keyManagementProvider, logger, isFetchSuccessful, err.Error(), lastFetchedTime, nil)
		return ctrl.Result{}, fmt.Errorf("Error fetching certificates in KMProvider %v with %v provider, error: %w", resource, keyManagementProvider.Spec.Type, err)
	}
	keymanagementprovider.SetCertificatesInMap(resource, certificates)
	isFetchSuccessful = true
	emptyErrorString := ""
	writeKMProviderStatus(ctx, r, &keyManagementProvider, logger, isFetchSuccessful, emptyErrorString, lastFetchedTime, certAttributes)

	logger.Infof("%v certificates fetched for key management provider %v", len(certificates), resource)

	// returning empty result and no error to indicate we’ve successfully reconciled this object
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KeyManagementProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred := predicate.GenerationChangedPredicate{}

	// status updates will trigger a reconcile event
	// if there are no changes to spec of CRD, this event should be filtered out by using the predicate
	// see more discussions at https://github.com/kubernetes-sigs/kubebuilder/issues/618
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1beta1.KeyManagementProvider{}).WithEventFilter(pred).
		Complete(r)
}

// specToKeyManagementProvider creates KeyManagementProviderProvider from  KeyManagementProviderSpec config
func specToKeyManagementProvider(spec configv1beta1.KeyManagementProviderSpec) (keymanagementprovider.KeyManagementProvider, error) {
	kmProviderConfig, err := rawToKeyManagementProviderConfig(spec.Parameters.Raw, spec.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to parse key management provider config: %w", err)
	}

	// TODO: add Version and Address to KeyManagementProviderSpec
	keyManagementProviderProvider, err := factory.CreateKeyManagementProviderFromConfig(kmProviderConfig, "0.1.0", c.GetDefaultPluginPath())
	if err != nil {
		return nil, fmt.Errorf("failed to create key management provider provider: %w", err)
	}

	return keyManagementProviderProvider, nil
}

// rawToKeyManagementProviderConfig converts raw json to KeyManagementProviderConfig
func rawToKeyManagementProviderConfig(raw []byte, keyManagamentSystemName string) (config.KeyManagementProviderConfig, error) {
	pluginConfig := config.KeyManagementProviderConfig{}

	if string(raw) == "" {
		return config.KeyManagementProviderConfig{}, fmt.Errorf("no key management provider parameters provided")
	}
	if err := json.Unmarshal(raw, &pluginConfig); err != nil {
		return config.KeyManagementProviderConfig{}, fmt.Errorf("unable to decode key management provider parameters.Raw: %s, err: %w", raw, err)
	}

	pluginConfig[types.Type] = keyManagamentSystemName

	return pluginConfig, nil
}

// writeKMProviderStatus updates the status of the key management provider resource
func writeKMProviderStatus(ctx context.Context, r client.StatusClient, keyManagementProvider *configv1beta1.KeyManagementProvider, logger *logrus.Entry, isSuccess bool, errorString string, operationTime metav1.Time, kmProviderStatus keymanagementprovider.KeyManagementProviderStatus) {
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
	if len(errorString) > maxBriefErrLength {
		briefErr = fmt.Sprintf("%s...", errorString[:maxBriefErrLength])
	}
	keyManagementProvider.Status.IsSuccess = false
	keyManagementProvider.Status.Error = errorString
	keyManagementProvider.Status.BriefError = briefErr
	keyManagementProvider.Status.LastFetchedTime = operationTime
}

// updateKMProviderSuccessStatus updates the key management provider status if status argument is non nil
// Success status includes last fetched time and other provider-specific properties
func updateKMProviderSuccessStatus(keyManagementProvider *configv1beta1.KeyManagementProvider, lastOperationTime *metav1.Time, kmProviderStatus keymanagementprovider.KeyManagementProviderStatus) {
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
