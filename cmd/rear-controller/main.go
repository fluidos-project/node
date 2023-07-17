/*
Copyright 2023.

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

package main

import (
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	advertisementv1alpha1 "fluidos.eu/node/api/advertisement/v1alpha1"
	reservationv1alpha1 "fluidos.eu/node/api/reservation/v1alpha1"
	contractmanager "fluidos.eu/node/pkg/rear-controller/contract-manager"
	discoverymanager "fluidos.eu/node/pkg/rear-controller/discovery-manager"
	gateway "fluidos.eu/node/pkg/rear-controller/gateway"
	"fluidos.eu/node/pkg/utils/flags"
	"fluidos.eu/node/pkg/utils/namings"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(advertisementv1alpha1.AddToScheme(scheme))
	utilruntime.Must(reservationv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	clientID, _ := namings.ForgePrefixClientID()

	// TODO: after the demo recover correct addresses (8080 and 8081)
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":7080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":7081", "The address the probe endpoint binds to.")
	flag.StringVar(&flags.DOMAIN, "domain", "fluidos.eu", "Domain name of rhw FLUIDOS node")
	flag.StringVar(&flags.IP_ADDR, "ip", "", "IP address of the FLUIDOS node")
	flag.StringVar(&flags.CLIENT_ID, "client-id", clientID, "Client ID related to the FLUIDOS node")
	// TODO: remember to delete this flag since the reservation is not created by the discovery manager
	flag.StringVar(&flags.RESERVATION_DEFAULT_NAMESPACE, "reservation-namespace", "default", "Namespace where the Reservation Custom Resources are created")
	flag.StringVar(&flags.DEFAULT_NAMESPACE, "default-namespace", "default", "Default namespace used by the FLUIDOS node")
	flag.StringVar(&flags.PC_DEFAULT_NAMESPACE, "pc-namespace", "default", "Default Namespace where the peering candidate CRs are created")
	flag.StringVar(&flags.RESOURCE_TYPE, "resources-types", "k8s-fluidos", "Type of the Flavour (for now we consider only k8s resources)")
	flag.StringVar(&flags.SERVER_ADDR, "server-addr", "http://localhost:14144/api", "Address of neighbour server used to discover other FLUIDOS nodes")
	flag.StringVar(&flags.SERVER_ADDRESSES[0], "server-address", flags.SERVER_ADDR, "Array of addresses of neighbour servers used to discover other FLUIDOS nodes")
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
		LeaderElectionID:       "efa8b828.fluidos.eu",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	gw := gateway.NewGateway(mgr.GetClient())

	if err = (&discoverymanager.DiscoveryReconciler{
		Client:  mgr.GetClient(),
		Scheme:  mgr.GetScheme(),
		Gateway: gw,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Discovery")
		os.Exit(1)
	}

	if err = (&contractmanager.ReservationReconciler{
		Client:  mgr.GetClient(),
		Scheme:  mgr.GetScheme(),
		Gateway: gw,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Reservation")
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

	// Start the HTTP server
	go func() {
		gw.StartHttpServer()
	}()

	// TODO: Uncomment this when the webhook is ready. For now it does not work (Ale)
	// pcv := discoverymanager.NewPCValidator(mgr.GetClient())

	// mgr.GetWebhookServer().Register("/validate/peeringcandidate", &webhook.Admission{Handler: pcv})

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
