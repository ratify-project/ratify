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

	configv1alpha1 "github.com/deislabs/ratify/api/v1alpha1"
	"github.com/deislabs/ratify/config"
	vr "github.com/deislabs/ratify/pkg/verifier"
	vc "github.com/deislabs/ratify/pkg/verifier/config"
	vf "github.com/deislabs/ratify/pkg/verifier/factory"
	"github.com/deislabs/ratify/pkg/verifier/types"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// VerifierReconciler reconciles a Verifier object
type VerifierReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var (
	// a map to track of active verifiers
	VerifierMap    = map[string]vr.ReferenceVerifier{}
	verifierLogger = log.Log
)

//+kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=verifiers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=verifiers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=verifiers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Verifier object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *VerifierReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	verifierLogger = log.FromContext(ctx)

	var verifier configv1alpha1.Verifier
	var resource = req.Name
	storeLogger.Info(fmt.Sprintf("reconciling verifier '%v'", resource))

	if err := r.Get(ctx, req.NamespacedName, &verifier); err != nil {

		if apierrors.IsNotFound(err) {
			verifierLogger.Info(fmt.Sprintf("delete event detected, removing verifier %v", resource))
			verifierRemove(resource)
		} else {
			verifierLogger.Error(err, "unable to fetch verifier")
		}

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := verifierAddOrReplace(verifier.Spec, resource); err != nil {
		verifierLogger.Error(err, "unable to create verifier from verifier crd")
		return ctrl.Result{}, err
	}

	// returning empty result and no error to indicate we’ve successfully reconciled this object
	return ctrl.Result{}, nil
}

// creates a verifier reference from CRD spec and add store to map
func verifierAddOrReplace(spec configv1alpha1.VerifierSpec, objectName string) error {

	verifierConfig, err := specToVerifierConfig(spec)

	if err != nil {
		verifierLogger.Error(err, "unable to convert crd specification to verifier config")
		return fmt.Errorf("unable to convert crd specification to verifier config, err: %q", err)
	}

	// verifier factory only support a single version of configuration today
	// when we support multi version verifier CRD, we will also pass in the corresponding config version so factory can create different version of the object
	verifierConfigVersion := "1.0.0" // TODO: move default values to defaulting webhook in the future #413
	if spec.Address == "" {
		spec.Address = config.GetDefaultPluginPath()
		verifierLogger.Info(fmt.Sprintf("Address was empty, setting to default path %v", spec.Address))
	}
	verifierReference, err := vf.CreateVerifierFromConfig(verifierConfig, verifierConfigVersion, []string{spec.Address})

	if err != nil || verifierReference == nil {
		verifierLogger.Error(err, "unable to create verifier from verifier config")
		return err
	}
	VerifierMap[objectName] = verifierReference
	verifierLogger.Info(fmt.Sprintf("verifier '%v' added to verifier map", verifierReference.Name()))

	return nil
}

// remove verifier from map
func verifierRemove(objectName string) {
	delete(VerifierMap, objectName)
}

// returns a verifier reference from spec
func specToVerifierConfig(verifierSpec configv1alpha1.VerifierSpec) (vc.VerifierConfig, error) {

	verifierConfig := vc.VerifierConfig{}

	if string(verifierSpec.Parameters.Raw) != "" {
		if err := json.Unmarshal(verifierSpec.Parameters.Raw, &verifierConfig); err != nil {
			verifierLogger.Error(err, "unable to decode verifier parameters", "Parameters.Raw", verifierSpec.Parameters.Raw)
			return vc.VerifierConfig{}, err
		}
	}

	verifierConfig[types.Name] = verifierSpec.Name
	verifierConfig[types.ArtifactTypes] = verifierSpec.ArtifactTypes

	return verifierConfig, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VerifierReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1alpha1.Verifier{}).
		Complete(r)
}
