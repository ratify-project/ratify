/*
Copyright 2022.

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
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	configv1alpha1 "github.com/deislabs/ratify/api/v1alpha1"
	"github.com/sirupsen/logrus"
)

// CertificateStoreReconciler reconciles a CertificateStore object
type CertificateStoreReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// This is a map containing Cert store configuration including name, tenantID, and cert object information
type CertStoreConfig map[string]interface{}

var (
	// a map between CertificateStore name to array of x509 certificates
	certificatesMap = map[string][]*x509.Certificate{}
)

//+kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=certificatestores,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=certificatestores/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=certificatestores/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CertificateStore object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *CertificateStoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logrus.WithContext(ctx)

	var resource = req.Name
	var certStore configv1alpha1.CertificateStore

	logger.Infof("reconciling certificate store '%v'", resource)

	if err := r.Get(ctx, req.NamespacedName, &certStore); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Infof("deletion detected, removing store %v", req.Name)
			//storeRemove(resource)
		} else {
			logger.Error(err, "unable to fetch certificate store")
		}

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Infof("meta data of the certStore %s", certStore.Name)
	logger.Infof("name of provider %s", certStore.Spec.Provider)

	// get the attribute map
	attributes, err := getCertStoreConfig(certStore.Spec)
	if err != nil {
		return ctrl.Result{}, err
	}

	// this can a new fetch or an update
	switch certStore.Spec.Provider {
	case "azurekeyvault":
		logrus.Infof("len %v", len(attributes))
		//var result = azurekeyvault.GetCertificates(attributes) returns
		//CertificatesMap[resource] = byteToCerts([][]bytes)
		//CertificatesMap[name] = result
	default:
		logger.Errorf("Unknown cert provider %s", certStore.Spec.Provider)
		//TODO: return a error
	}

	// returning empty result and no error to indicate we’ve successfully reconciled this object
	return ctrl.Result{}, nil
}

// converts array of cert content to x509 certificate
func byteToCerts(certificates [][]byte) ([]*x509.Certificate, error) {
	r := []*x509.Certificate{}
	for _, c := range certificates {
		block, _ := pem.Decode(c)
		if block == nil {
			return nil, fmt.Errorf("Failed to decode certificate")
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}

		r = append(r, cert)
	}

	return r, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CertificateStoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1alpha1.CertificateStore{}).
		Complete(r)
}

func getCertStoreConfig(spec configv1alpha1.CertificateStoreSpec) (CertStoreConfig, error) {
	attributes := CertStoreConfig{}

	if string(spec.Parameters.Raw) != "" {
		if err := json.Unmarshal(spec.Parameters.Raw, &attributes); err != nil {
			logrus.Error(err, "unable to decode cert store parameters", "Parameters.Raw", spec.Parameters.Raw)
			return attributes, err
		}
	}
	return attributes, nil
}

func GetCertificatesMap() map[string][]*x509.Certificate {
	return certificatesMap
}
