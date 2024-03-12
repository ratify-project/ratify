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
	"context"
	"crypto/x509"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/go-logr/logr"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"github.com/deislabs/ratify/config"
	"github.com/deislabs/ratify/httpserver"
	"github.com/deislabs/ratify/pkg/featureflag"
	"github.com/deislabs/ratify/pkg/policyprovider"
	_ "github.com/deislabs/ratify/pkg/policyprovider/configpolicy" // register config policy provider
	_ "github.com/deislabs/ratify/pkg/policyprovider/regopolicy"   // register rego policy provider
	_ "github.com/deislabs/ratify/pkg/referrerstore/oras"          // register ORAS referrer store
	"github.com/deislabs/ratify/pkg/utils"
	_ "github.com/deislabs/ratify/pkg/verifier/notation" // register notation verifier
	"github.com/open-policy-agent/cert-controller/pkg/rotator"
	"github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // import additional authentication methods

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	configv1alpha1 "github.com/deislabs/ratify/api/v1alpha1"
	configv1beta1 "github.com/deislabs/ratify/api/v1beta1"
	"github.com/deislabs/ratify/pkg/controllers"
	ef "github.com/deislabs/ratify/pkg/executor/core"
	"github.com/deislabs/ratify/pkg/referrerstore"
	vr "github.com/deislabs/ratify/pkg/verifier"
	//+kubebuilder:scaffold:imports
)

const (
	caOrganization = "Ratify"
	certDir        = "/usr/local/tls"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = logrus.WithField("name", "setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(configv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
	utilruntime.Must(configv1beta1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func StartServer(httpServerAddress, configFilePath, certDirectory, caCertFile string, cacheTTL time.Duration, metricsEnabled bool, metricsType string, metricsPort int, certRotatorReady chan struct{}) {
	logrus.Info("initializing executor with config file at default config path")

	cf, err := config.Load(configFilePath)
	if err != nil {
		logrus.Errorf("server start failed %v", fmt.Errorf("error loading config %w", err))
		os.Exit(1)
	}

	// initialize server
	server, err := httpserver.NewServer(context.Background(), httpServerAddress, func() *ef.Executor {
		var activeVerifiers []vr.ReferenceVerifier
		var activeStores []referrerstore.ReferrerStore
		var activePolicyEnforcer policyprovider.PolicyProvider

		// check if there are active verifiers from crd controller
		if len(controllers.VerifierMap) > 0 {
			for _, value := range controllers.VerifierMap {
				activeVerifiers = append(activeVerifiers, value)
			}
		}

		// check if there are active stores from crd controller
		if len(controllers.StoreMap) > 0 {
			for _, value := range controllers.StoreMap {
				activeStores = append(activeStores, value)
			}
		}

		if !controllers.ActivePolicy.IsEmpty() {
			activePolicyEnforcer = controllers.ActivePolicy.Enforcer
		}

		// return executor with latest configuration
		executor := ef.Executor{
			Verifiers:      activeVerifiers,
			ReferrerStores: activeStores,
			PolicyEnforcer: activePolicyEnforcer,
			Config:         &cf.ExecutorConfig,
		}
		return &executor
	}, certDirectory, caCertFile, cacheTTL, metricsEnabled, metricsType, metricsPort)

	if err != nil {
		logrus.Errorf("initialize server failed with error %v, exiting..", err)
		os.Exit(1)
	}
	logrus.Infof("starting server at" + httpServerAddress)
	if err := server.Run(certRotatorReady); err != nil {
		logrus.Errorf("starting server failed with error %v, exiting..", err)
		os.Exit(1)
	}
}

func StartManager(certRotatorReady chan struct{}, probeAddr string) {
	var metricsAddr string
	var enableLeaderElection bool

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	logrusSink := controllers.NewLogrusSink(logrus.StandardLogger())
	ctrl.SetLogger(logr.New(logrusSink))
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "1a306109.github.com/deislabs/ratify",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	setupLog.Debugf("setting up probeAddr at %s", probeAddr)

	// Make sure certs are generated and valid if cert rotation is enabled.
	if featureflag.CertRotation.Enabled {
		// Make sure TLS cert watcher is already set up.
		if certRotatorReady == nil {
			setupLog.Error(err, "to use cert rotation, you must provide a channel to signal when the TLS watcher is ready")
			os.Exit(1)
		}
		setupLog.Info("setting up cert rotation")

		keyUsages := []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
		webhooks := []rotator.WebhookInfo{
			{
				Name: "ratify-mutation-provider",
				Type: rotator.ExternalDataProvider,
			},
			{
				Name: "ratify-provider",
				Type: rotator.ExternalDataProvider,
			},
		}
		namespace := utils.GetNamespace()
		serviceName := utils.GetServiceName()

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
			setupLog.Error(err, "Unable to set up cert rotation")
			os.Exit(1)
		}
	} else {
		close(certRotatorReady)
	}

	if err = (&controllers.VerifierReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Verifier")
		os.Exit(1)
	}
	if err = (&controllers.StoreReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Store")
		os.Exit(1)
	}
	if err = (&controllers.CertificateStoreReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Certificate Store")
		os.Exit(1)
	}
	if err = (&controllers.PolicyReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Policy")
		os.Exit(1)
	}
	if err = (&controllers.KeyManagementProviderReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Key Management Provider")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
