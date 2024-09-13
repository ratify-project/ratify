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

package namespaceresource

import (
	"context"
	"fmt"

	_ "github.com/ratify-project/ratify/pkg/keymanagementprovider/azurekeyvault" // register azure key vault key management provider
	_ "github.com/ratify-project/ratify/pkg/keymanagementprovider/inline"        // register inline key management provider
	"github.com/ratify-project/ratify/pkg/keymanagementprovider/refresh"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	configv1beta1 "github.com/ratify-project/ratify/api/v1beta1"
)

// KeyManagementProviderReconciler reconciles a KeyManagementProvider object
type KeyManagementProviderReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *KeyManagementProviderReconciler) ReconcileWithConfig(ctx context.Context, config map[string]interface{}) (ctrl.Result, error) {
	refresher, err := refresh.CreateRefresherFromConfig(config)
	if err != nil {
		return ctrl.Result{}, err
	}
	err = refresher.Refresh(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	result, ok := refresher.GetResult().(ctrl.Result)
	if !ok {
		return ctrl.Result{}, fmt.Errorf("unexpected type returned from GetResult: %T", refresher.GetResult())
	}
	return result, nil
}

// +kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=namespacedkeymanagementproviders,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=namespacedkeymanagementproviders/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=namespacedkeymanagementproviders/finalizers,verbs=update
func (r *KeyManagementProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	refresherConfig := map[string]interface{}{
		"type":    refresh.KubeRefresherNamespacedType,
		"client":  r.Client,
		"request": req,
	}
	return r.ReconcileWithConfig(ctx, refresherConfig)
}

// SetupWithManager sets up the controller with the Manager.
func (r *KeyManagementProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred := predicate.GenerationChangedPredicate{}

	// status updates will trigger a reconcile event
	// if there are no changes to spec of CRD, this event should be filtered out by using the predicate
	// see more discussions at https://github.com/kubernetes-sigs/kubebuilder/issues/618
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1beta1.NamespacedKeyManagementProvider{}).WithEventFilter(pred).
		Complete(r)
}
