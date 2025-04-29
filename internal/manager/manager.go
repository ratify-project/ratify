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

package manager

import (
	"crypto/x509"
	"fmt"
	"os"

	"github.com/open-policy-agent/cert-controller/pkg/rotator"
	"github.com/ratify-project/ratify/v2/internal/pod"
	"k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	caOrganization = "Ratify"
	certDir        = "/usr/local/tls"
)

func StartManager(certRotatorReady chan struct{}) {
	log := ctrl.Log.WithName("ratify-manager")
	if certRotatorReady == nil {
		log.Info("cert rotator is not enabled")
		return
	}

	log.Info("setting up cert rotation")
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{})
	if err != nil {
		log.Error(err, "could not create ratify manager")
		os.Exit(1)
	}

	keyUsages := []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	webhooks := []rotator.WebhookInfo{
		{
			Name: "ratify-gatekeeper-provider",
			Type: rotator.ExternalDataProvider,
		},
	}
	namespace := pod.GetNamespace()
	serviceName := pod.GetServiceName()

	if err := rotator.AddRotator(mgr, &rotator.CertRotator{
		SecretKey: types.NamespacedName{
			Namespace: namespace,
			Name:      fmt.Sprintf("%s-tls", serviceName),
		},
		CertDir:        certDir,
		CAName:         fmt.Sprintf("%s.%s", serviceName, namespace),
		CAOrganization: caOrganization,
		DNSName:        fmt.Sprintf("%s.%s", serviceName, namespace),
		IsReady:        certRotatorReady,
		Webhooks:       webhooks,
		ExtKeyUsages:   &keyUsages,
	}); err != nil {
		log.Error(err, "unable to set up cert rotation")
		os.Exit(1)
	}

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "could not start manager")
		os.Exit(1)
	}
}
