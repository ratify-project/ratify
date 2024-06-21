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

package clusterresource

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"

	"github.com/ratify-project/ratify/internal/constants"
	_ "github.com/ratify-project/ratify/pkg/keymanagementprovider/azurekeyvault" // register azure key vault key management provider
	_ "github.com/ratify-project/ratify/pkg/keymanagementprovider/inline"        // register inline key management provider
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	configv1beta1 "github.com/ratify-project/ratify/api/v1beta1"
	cutils "github.com/ratify-project/ratify/pkg/controllers/utils"
	kmp "github.com/ratify-project/ratify/pkg/keymanagementprovider"
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

	var resource = req.Name
	var keyManagementProvider configv1beta1.KeyManagementProvider

	logger.Infof("reconciling cluster key management provider '%v'", resource)

	if err := r.Get(ctx, req.NamespacedName, &keyManagementProvider); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Infof("deletion detected, removing key management provider %v", resource)
			kmp.DeleteCertificatesFromMap(resource)
			kmp.DeleteKeysFromMap(resource)
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
		// Note: for backwards compatibility in upgrade scenarios, Ratify will only log a warning statement.
		logger.Warn("Certificate Store already exists. Key management provider and certificate store should not be configured together. Please migrate to key management provider and delete certificate store.")
	}

	provider, err := cutils.SpecToKeyManagementProvider(keyManagementProvider.Spec.Parameters.Raw, keyManagementProvider.Spec.Type)
	if err != nil {
		writeKMProviderStatus(ctx, r, &keyManagementProvider, logger, isFetchSuccessful, err.Error(), lastFetchedTime, nil)
		return ctrl.Result{}, err
	}

	// fetch certificates and store in map
	certificates, certAttributes, err := provider.GetCertificates(ctx)
	if err != nil {
		writeKMProviderStatus(ctx, r, &keyManagementProvider, logger, isFetchSuccessful, err.Error(), lastFetchedTime, nil)
		return ctrl.Result{}, fmt.Errorf("Error fetching certificates in KMProvider %v with %v provider, error: %w", resource, keyManagementProvider.Spec.Type, err)
	}

	// fetch keys and store in map
	keys, keyAttributes, err := provider.GetKeys(ctx)
	if err != nil {
		writeKMProviderStatus(ctx, r, &keyManagementProvider, logger, isFetchSuccessful, err.Error(), lastFetchedTime, nil)
		return ctrl.Result{}, fmt.Errorf("Error fetching keys in KMProvider %v with %v provider, error: %w", resource, keyManagementProvider.Spec.Type, err)
	}
	kmp.SetCertificatesInMap(resource, certificates)
	kmp.SetKeysInMap(resource, keyManagementProvider.Spec.Type, keys)
	// merge certificates and keys status into one
	maps.Copy(keyAttributes, certAttributes)
	isFetchSuccessful = true
	emptyErrorString := ""
	writeKMProviderStatus(ctx, r, &keyManagementProvider, logger, isFetchSuccessful, emptyErrorString, lastFetchedTime, keyAttributes)

	logger.Infof("%v certificate(s) & %v key(s) fetched for key management provider %v", len(certificates), len(keys), resource)

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

// writeKMProviderStatus updates the status of the key management provider resource
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
