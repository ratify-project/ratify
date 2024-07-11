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

	"github.com/ratify-project/ratify/internal/constants"
	_ "github.com/ratify-project/ratify/pkg/keymanagementprovider/azurekeyvault" // register azure key vault key management provider
	_ "github.com/ratify-project/ratify/pkg/keymanagementprovider/inline"        // register inline key management provider
	"github.com/ratify-project/ratify/pkg/keymanagementprovider/refresh"         // register inline key management provider
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	configv1beta1 "github.com/ratify-project/ratify/api/v1beta1"
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
	kr := refresh.KubeRefresher{
		Client:  r.Client,
		Request: req,
	}

	// check if kr.client is nil
	if kr.Client == nil {
		return ctrl.Result{}, fmt.Errorf("client is nil")
	}

	err := kr.Refresh(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	return kr.Result, nil
}

// TODO: delete helpers, moved to kubeRefresh.go
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
