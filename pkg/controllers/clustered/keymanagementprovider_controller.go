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

package clustered

import (
	"context"
	"encoding/json"
	"fmt"

	configv1beta1 "github.com/deislabs/ratify/api/v1beta1"
	re "github.com/deislabs/ratify/errors"
	"github.com/deislabs/ratify/internal/constants"
	"github.com/deislabs/ratify/pkg/controllers"
	"github.com/deislabs/ratify/pkg/controllers/utils"
	"github.com/deislabs/ratify/pkg/keymanagementprovider"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// KeyManagementProviderReconciler reconciles a KeyManagementProvider object
type KeyManagementProviderReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=clusterkeymanagementproviders,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=clusterkeymanagementproviders/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=clusterkeymanagementproviders/finalizers,verbs=update
func (r *KeyManagementProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logrus.WithContext(ctx)

	var resource = req.NamespacedName.String()
	var keyManagementProvider configv1beta1.ClusterKeyManagementProvider

	logger.Infof("reconciling cluster key management provider '%v'", resource)

	if err := r.Get(ctx, req.NamespacedName, &keyManagementProvider); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Infof("deletion detected, removing key management provider %v", resource)
			controllers.KMPMap.DeleteStore(constants.EmptyNamespace, resource)
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
	provider, err := utils.GetKeyManagementProvider(keyManagementProvider.Spec.Parameters.Raw, keyManagementProvider.Spec.Type)
	if err != nil {
		writeKMProviderStatus(ctx, r, &keyManagementProvider, logger, isFetchSuccessful, err.Error(), lastFetchedTime, nil)
		return ctrl.Result{}, err
	}

	certificates, certAttributes, err := provider.GetCertificates(ctx)
	if err != nil {
		writeKMProviderStatus(ctx, r, &keyManagementProvider, logger, isFetchSuccessful, err.Error(), lastFetchedTime, nil)
		return ctrl.Result{}, fmt.Errorf("Error fetching certificates in KMProvider %v with %v provider, error: %w", resource, keyManagementProvider.Spec.Type, err)
	}
	controllers.KMPMap.AddStore(constants.EmptyNamespace, resource, certificates)
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

// writeKMProviderStatus updates the status of the key management provider resource
func writeKMProviderStatus(ctx context.Context, r client.StatusClient, keyManagementProvider *configv1beta1.ClusterKeyManagementProvider, logger *logrus.Entry, isSuccess bool, errorString string, operationTime metav1.Time, kmProviderStatus keymanagementprovider.KeyManagementProviderStatus) {
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
func updateKMProviderErrorStatus(keyManagementProvider *configv1beta1.ClusterKeyManagementProvider, errorString string, operationTime *metav1.Time) {
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
func updateKMProviderSuccessStatus(keyManagementProvider *configv1beta1.ClusterKeyManagementProvider, lastOperationTime *metav1.Time, kmProviderStatus keymanagementprovider.KeyManagementProviderStatus) {
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
