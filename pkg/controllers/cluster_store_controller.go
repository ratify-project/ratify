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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	configv1beta1 "github.com/deislabs/ratify/api/v1beta1"
	"github.com/deislabs/ratify/config"
	"github.com/deislabs/ratify/pkg/referrerstore"
	rc "github.com/deislabs/ratify/pkg/referrerstore/config"
	sf "github.com/deislabs/ratify/pkg/referrerstore/factory"
	"github.com/deislabs/ratify/pkg/referrerstore/types"
	"github.com/sirupsen/logrus"
)

// ClusterStoreReconciler reconciles a Store object
type ClusterStoreReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var (
	// a map to track active cluster stores
	ClusterStoreMap = map[string]referrerstore.ReferrerStore{}
)

//+kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=clusterstores,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=clusterstores/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=clusterstores/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *ClusterStoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	storeLogger := logrus.WithContext(ctx)

	var store configv1beta1.ClusterStore
	var resource = req.Name
	storeLogger.Infof("reconciling store '%v'", resource)

	if err := r.Get(ctx, req.NamespacedName, &store); err != nil {
		if apierrors.IsNotFound(err) {
			storeLogger.Infof("deletion detected, removing store %v", req.Name)
			clusterStoreRemove(resource)
		} else {
			storeLogger.Error(err, "unable to fetch store")
		}

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := clusterStoreAddOrReplace(store.Spec, resource); err != nil {
		storeLogger.Error(err, "unable to create store from store crd")
		return ctrl.Result{}, err
	}

	// returning empty result and no error to indicate we’ve successfully reconciled this object
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterStoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1beta1.Store{}).
		Complete(r)
}

// Creates a store reference from CRD spec and add store to map
func clusterStoreAddOrReplace(spec configv1beta1.ClusterStoreSpec, fullname string) error {
	storeConfig, err := clusterSpecToStoreConfig(spec)
	if err != nil {
		return fmt.Errorf("unable to convert store spec to store config, err: %w", err)
	}

	// if the default version is not suitable, the plugin configuration should specify the desired version
	if len(spec.Version) == 0 {
		spec.Version = config.GetDefaultPluginVersion()
		logrus.Infof("Version was empty, setting to default version: %v", spec.Version)
	}

	if spec.Address == "" {
		spec.Address = config.GetDefaultPluginPath()
		logrus.Infof("Address was empty, setting to default path %v", spec.Address)
	}
	storeReference, err := sf.CreateStoreFromConfig(storeConfig, spec.Version, []string{spec.Address})

	if err != nil || storeReference == nil {
		logrus.Error(err, "store factory failed to create store from store config")
		return fmt.Errorf("store factory failed to create store from store config, err: %w", err)
	}

	ClusterStoreMap[fullname] = storeReference
	logrus.Infof("store '%v' added to store map", storeReference.Name())

	return nil
}

// Remove store from map
func clusterStoreRemove(resourceName string) {
	delete(ClusterStoreMap, resourceName)
}

// Returns a store reference from spec
func clusterSpecToStoreConfig(storeSpec configv1beta1.ClusterStoreSpec) (rc.StorePluginConfig, error) {
	storeConfig := rc.StorePluginConfig{}

	if string(storeSpec.Parameters.Raw) != "" {
		if err := json.Unmarshal(storeSpec.Parameters.Raw, &storeConfig); err != nil {
			logrus.Error(err, "unable to decode store parameters", "Parameters.Raw", storeSpec.Parameters.Raw)
			return rc.StorePluginConfig{}, err
		}
	}
	storeConfig[types.Name] = storeSpec.Name
	if storeSpec.Source != nil {
		storeConfig[types.Source] = storeSpec.Source
	}

	return storeConfig, nil
}
