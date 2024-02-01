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
	"os"

	configv1beta1 "github.com/deislabs/ratify/api/v1beta1"
	"github.com/deislabs/ratify/config"
	re "github.com/deislabs/ratify/errors"
	"github.com/deislabs/ratify/pkg/utils"
	vr "github.com/deislabs/ratify/pkg/verifier"
	vc "github.com/deislabs/ratify/pkg/verifier/config"
	vf "github.com/deislabs/ratify/pkg/verifier/factory"
	"github.com/deislabs/ratify/pkg/verifier/types"

	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// VerifierReconciler reconciles a Verifier object
type VerifierReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var (
	// a map to track of active verifiers
	VerifierMap = map[string]vr.ReferenceVerifier{}
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
	verifierLogger := logrus.WithContext(ctx)

	var verifier configv1beta1.Verifier
	var resource = req.Name

	verifierLogger.Infof("reconciling verifier '%v'", resource)

	if err := r.Get(ctx, req.NamespacedName, &verifier); err != nil {
		if apierrors.IsNotFound(err) {
			verifierLogger.Infof("delete event detected, removing verifier %v", resource)
			verifierRemove(resource)
		} else {
			verifierLogger.Error(err, "unable to fetch verifier")
		}

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	namespace, err := getCertStoreNamespace(req.Namespace)
	if err != nil {
		verifierLogger.Error(err, "unable to get default namespace for certstore specified in verifier crd")
		return ctrl.Result{}, err
	}

	if err = verifierAddOrReplace(verifier.Spec, resource, namespace); err != nil {
		verifierLogger.Error(err, "unable to create verifier from verifier crd")
		writeVerifierStatus(ctx, r, &verifier, verifierLogger, false, err.Error())
		return ctrl.Result{}, err
	}

	writeVerifierStatus(ctx, r, &verifier, verifierLogger, true, "")

	// returning empty result and no error to indicate we’ve successfully reconciled this object
	return ctrl.Result{}, nil
}

// creates a verifier reference from CRD spec and add store to map
func verifierAddOrReplace(spec configv1beta1.VerifierSpec, objectName string, namespace string) error {
	verifierConfig, err := specToVerifierConfig(spec, objectName)
	if err != nil {
		logrus.Error(err, "unable to convert crd specification to verifier config")
		return fmt.Errorf("unable to convert crd specification to verifier config, err: %w", err)
	}

	if len(spec.Version) == 0 {
		spec.Version = config.GetDefaultPluginVersion()
		logrus.Infof("Version was empty, setting to default version: %v", spec.Version)
	}

	if spec.Address == "" {
		spec.Address = config.GetDefaultPluginPath()
		logrus.Infof("Address was empty, setting to default path: %v", spec.Address)
	}

	referenceVerifier, err := vf.CreateVerifierFromConfig(verifierConfig, spec.Version, []string{spec.Address}, namespace)

	if err != nil || referenceVerifier == nil {
		logrus.Error(err, "unable to create verifier from verifier config")
		return err
	}
	VerifierMap[objectName] = referenceVerifier
	logrus.Infof("verifier '%v' added to verifier map", referenceVerifier.Name())

	return nil
}

// remove verifier from map
func verifierRemove(objectName string) {
	delete(VerifierMap, objectName)
}

// returns a verifier reference from spec
func specToVerifierConfig(verifierSpec configv1beta1.VerifierSpec, verifierName string) (vc.VerifierConfig, error) {
	verifierConfig := vc.VerifierConfig{}

	if string(verifierSpec.Parameters.Raw) != "" {
		if err := json.Unmarshal(verifierSpec.Parameters.Raw, &verifierConfig); err != nil {
			logrus.Error(err, "unable to decode verifier parameters", "Parameters.Raw", verifierSpec.Parameters.Raw)
			return vc.VerifierConfig{}, err
		}
	}
	verifierConfig[types.Name] = verifierName
	verifierConfig[types.Type] = verifierSpec.Name
	verifierConfig[types.ArtifactTypes] = verifierSpec.ArtifactTypes
	if verifierSpec.Source != nil {
		verifierConfig[types.Source] = verifierSpec.Source
	}

	return verifierConfig, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VerifierReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1beta1.Verifier{}).
		Complete(r)
}

// Historically certStore defined in trust policy only contains name which means the CertStore cannot be uniquely identified
// If verifierNamespace is not empty, this method returns the default cert store namespace else returns the ratify deployed namespace
func getCertStoreNamespace(verifierNamespace string) (string, error) {
	// first, check if we can use the verifier namespace as the cert store namespace
	if verifierNamespace != "" {
		return verifierNamespace, nil
	}

	// next, return the ratify deployed namespace
	ns, found := os.LookupEnv(utils.RatifyNamespaceEnvVar)
	if !found {
		return "", re.ErrorCodeEnvNotSet.WithComponentType(re.Verifier).WithDetail(fmt.Sprintf("environment variable %s not set", utils.RatifyNamespaceEnvVar))
	}

	return ns, nil
}

func writeVerifierStatus(ctx context.Context, r client.StatusClient, verifier *configv1beta1.Verifier, logger *logrus.Entry, isSuccess bool, errorString string) {
	if isSuccess {
		verifier.Status.IsSuccess = true
		verifier.Status.Error = ""
		verifier.Status.BriefError = ""
	} else {
		verifier.Status.IsSuccess = false
		verifier.Status.Error = errorString
		if len(errorString) > maxBriefErrLength {
			verifier.Status.BriefError = fmt.Sprintf("%s...", errorString[:maxBriefErrLength])
		}
	}

	if statusErr := r.Status().Update(ctx, verifier); statusErr != nil {
		logger.Error(statusErr, ",unable to update verifier status")
	}
}
