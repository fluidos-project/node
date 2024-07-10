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
	"flag"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog/v2"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	localresourcemanager "github.com/fluidos-project/node/pkg/local-resource-manager"
	"github.com/fluidos-project/node/pkg/utils/flags"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(metricsv1beta1.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(nodecorev1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&flags.AMOUNT, "amount", "", "Amount of money set for the flavors of this node")
	flag.StringVar(&flags.CURRENCY, "currency", "", "Currency of the money set for the flavors of this node")
	flag.StringVar(&flags.PERIOD, "period", "", "Period set for the flavors of this node")
	flag.StringVar(&flags.ResourceType, "resources-types", "k8s-fluidos", "Type of the Flavour related to k8s resources")
	flag.StringVar(&flags.CPUMin, "cpu-min", "0", "Minimum CPU value")
	flag.StringVar(&flags.MemoryMin, "memory-min", "0", "Minimum memory value")
	flag.StringVar(&flags.PodsMin, "pods-min", "0", "Minimum Pods value")
	flag.StringVar(&flags.CPUStep, "cpu-step", "0", "CPU step value")
	flag.StringVar(&flags.MemoryStep, "memory-step", "0", "Memory step value")
	flag.StringVar(&flags.PodsStep, "pods-step", "0", "Pods step value")
	flag.Int64Var(&flags.MinCount, "min-count", 0, "Minimum number of flavors")
	flag.Int64Var(&flags.MaxCount, "max-count", 0, "Maximum number of flavors")
	flag.StringVar(&flags.ResourceNodeLabel, "node-resource-label", "node-role.fluidos.eu/resources",
		"Label used to filter the k8s nodes from which create flavors")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	enableWH := flag.Bool("enable-webhooks", true, "Enable webhooks server")
	enableAutoDiscovery := flag.Bool("enable-auto-discovery", true, "Enable auto discovery of the node resources")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	cfg := ctrl.GetConfigOrDie()
	cl, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		setupLog.Error(err, "Unable to create client")
		os.Exit(1)
	}

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
		LeaderElectionID:       "c7b7b7b7.fluidos.eu",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Print something abou the mgr
	setupLog.Info("Manager started", "manager", mgr)

	// Register the controller
	if err = (&localresourcemanager.NodeReconciler{
		Client:              cl,
		Scheme:              mgr.GetScheme(),
		EnableAutoDiscovery: *enableAutoDiscovery,
		WebhookServer:       webhookServer,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Node")
		os.Exit(1)
	}

	if *enableWH {
		// Register Flavor webhook
		klog.Info("Registering webhooks to the manager")
		if err = (&nodecorev1alpha1.Flavor{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Flavor")
			os.Exit(1)
		}
	} else {
		setupLog.Info("Webhooks are disabled")
	}

	// Register health checks
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

	setupLog.Info("Starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
