// Copyright 2022-2024 FLUIDOS Project
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	liqodiscovery "github.com/liqotech/liqo/apis/discovery/v1alpha1"
	liqooffloading "github.com/liqotech/liqo/apis/offloading/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	advertisementv1alpha1 "github.com/fluidos-project/node/apis/advertisement/v1alpha1"
	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	reservationv1alpha1 "github.com/fluidos-project/node/apis/reservation/v1alpha1"
	rearmanager "github.com/fluidos-project/node/pkg/rear-manager"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(nodecorev1alpha1.AddToScheme(scheme))
	utilruntime.Must(advertisementv1alpha1.AddToScheme(scheme))
	utilruntime.Must(reservationv1alpha1.AddToScheme(scheme))
	utilruntime.Must(liqodiscovery.AddToScheme(scheme))
	utilruntime.Must(liqooffloading.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	enableWH := flag.Bool("enable-webhooks", true, "Enable webhooks server")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	var webhookServer webhook.Server

	if *enableWH {
		webhookServer = webhook.NewServer(webhook.Options{Port: 9443})
	} else {
		setupLog.Info("Webhooks are disabled")
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "586b6b69.fluidos.eu",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	cache := mgr.GetCache()

	// Index the RemoteClusterID field of the Allocation CRD
	indexFuncAllocation := func(obj client.Object) []string {
		allocation := obj.(*nodecorev1alpha1.Allocation)
		return []string{allocation.Name}
	}

	if err := cache.IndexField(context.Background(), &nodecorev1alpha1.Allocation{}, "spec.remoteClusterID", indexFuncAllocation); err != nil {
		setupLog.Error(err, "unable to create index for field", "field", "spec.remoteClusterID")
		os.Exit(1)
	}

	if err = (&rearmanager.SolverReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Solver")
		os.Exit(1)
	}

	if err = (&rearmanager.AllocationReconciler{
		Client:  mgr.GetClient(),
		Scheme:  mgr.GetScheme(),
		Manager: mgr,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Allocation")
		os.Exit(1)
	}

	if *enableWH {
		// Register Solver webhook
		setupLog.Info("Registering webhooks to the manager")
		if err = (&nodecorev1alpha1.Solver{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Flavor")
			os.Exit(1)
		}
	} else {
		setupLog.Info("Webhooks are disabled")
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

	if err := mgr.AddHealthzCheck("webhook", webhookServer.StartedChecker()); err != nil {
		setupLog.Error(err, "unable to set up webhook health check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
