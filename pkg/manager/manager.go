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
	"flag"
	"os"

	_ "github.com/deislabs/ratify/pkg/policyprovider/configpolicy"
	_ "github.com/deislabs/ratify/pkg/referrerstore/oras"
	_ "github.com/deislabs/ratify/pkg/verifier/notaryv2"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"github.com/deislabs/ratify/config"
	"github.com/deislabs/ratify/httpserver"
	_ "github.com/deislabs/ratify/pkg/policyprovider/configpolicy"
	_ "github.com/deislabs/ratify/pkg/referrerstore/oras"
	_ "github.com/deislabs/ratify/pkg/verifier/notaryv2"
	"github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/deislabs/ratify/pkg/verifier"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	configv1alpha1 "github.com/deislabs/ratify/api/v1alpha1"
	"github.com/deislabs/ratify/pkg/controllers"
	ef "github.com/deislabs/ratify/pkg/executor/core"
	"github.com/deislabs/ratify/pkg/policyprovider"
	"github.com/deislabs/ratify/pkg/referrerstore"
	vr "github.com/deislabs/ratify/pkg/verifier"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")

	configStores    []referrerstore.ReferrerStore
	configVerifiers []verifier.ReferenceVerifier
	policy          policyprovider.PolicyProvider
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(configv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func StartServer(httpServerAddress string, configFilePath string, certDirectory string, caCertFile string) {

	logrus.Infof("initializing executor with config file at default config path")

	cf, err := config.Load(configFilePath)

	configStores, configVerifiers, policy, err := config.CreateFromConfig(cf)

	if err != nil {
		logrus.Warnf("error initializing from config %v", err)
		os.Exit(1)
	}

	// initialize server
	server, err := httpserver.NewServer(context.Background(), httpServerAddress, func() *ef.Executor {

		var activeVerifiers []vr.ReferenceVerifier
		var activeStores []referrerstore.ReferrerStore

		// check if there are active verifiers from crd controller
		// else use verifiers from configuration
		if len(controllers.VerifierMap) > 0 {
			for _, value := range controllers.VerifierMap {
				activeVerifiers = append(activeVerifiers, value)
			}
		} else {
			activeVerifiers = configVerifiers
		}

		// check if there are active stores from crd controller
		// else use stores from configuration
		if len(controllers.StoreMap) > 0 {
			for _, value := range controllers.StoreMap {
				activeStores = append(activeStores, value)
			}
		} else {
			activeStores = configStores
		}

		// return executor with latest configuration
		executor := ef.Executor{
			Verifiers:      activeVerifiers,
			ReferrerStores: activeStores,
			PolicyEnforcer: policy,
			Config:         &cf.ExecutorConfig,
		}
		return &executor
	}, certDirectory, caCertFile)

	if err != nil {
		os.Exit(1)
	}
	logrus.Infof("starting server at" + httpServerAddress)
	if err := server.Run(); err != nil {
		os.Exit(1)
	}
}

func StartManager() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

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
