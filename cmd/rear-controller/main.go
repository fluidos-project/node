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
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	advertisementv1alpha1 "github.com/fluidos-project/node/apis/advertisement/v1alpha1"
	networkv1alpha1 "github.com/fluidos-project/node/apis/network/v1alpha1"
	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	reservationv1alpha1 "github.com/fluidos-project/node/apis/reservation/v1alpha1"
	contractmanager "github.com/fluidos-project/node/pkg/rear-controller/contract-manager"
	discoverymanager "github.com/fluidos-project/node/pkg/rear-controller/discovery-manager"
	gateway "github.com/fluidos-project/node/pkg/rear-controller/gateway"
	"github.com/fluidos-project/node/pkg/rear-controller/grpc"
	"github.com/fluidos-project/node/pkg/utils/flags"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(advertisementv1alpha1.AddToScheme(scheme))
	utilruntime.Must(reservationv1alpha1.AddToScheme(scheme))
	utilruntime.Must(nodecorev1alpha1.AddToScheme(scheme))
	utilruntime.Must(networkv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&flags.GRPCPort, "grpc-port", "2710", "Port of the HTTP server")
	flag.StringVar(&flags.HTTPPort, "http-port", "3004", "Port of the HTTP server")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
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
		LeaderElectionID:       "efa8b828.fluidos.eu",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	cache := mgr.GetCache()

	// Index the TransactionID field of the Contract CRD
	indexFuncTransaction := func(obj client.Object) []string {
		contract := obj.(*reservationv1alpha1.Contract)
		return []string{contract.Spec.TransactionID}
	}

	// Index the ClusterID field of the Contract CRD
	indexFuncClusterID := func(obj client.Object) []string {
		contract := obj.(*reservationv1alpha1.Contract)
		return []string{contract.Spec.BuyerClusterID}
	}

	indexFuncBuyerLiqoID := func(obj client.Object) []string {
		contract := obj.(*reservationv1alpha1.Contract)
		if contract.Spec.Buyer.AdditionalInformation == nil {
			return []string{}
		}
		return []string{contract.Spec.Buyer.AdditionalInformation.LiqoID}
	}

	indexFuncSellerLiqoID := func(obj client.Object) []string {
		contract := obj.(*reservationv1alpha1.Contract)
		if contract.Spec.Seller.AdditionalInformation == nil {
			return []string{}
		}
		return []string{contract.Spec.Seller.AdditionalInformation.LiqoID}
	}

	if err := cache.IndexField(
		context.Background(),
		&reservationv1alpha1.Contract{},
		"spec.transactionID",
		indexFuncTransaction,
	); err != nil {
		setupLog.Error(err, "unable to create index for field", "field", "spec.transactionID")
		os.Exit(1)
	}

	if err := cache.IndexField(
		context.Background(),
		&reservationv1alpha1.Contract{},
		"spec.buyerClusterID",
		indexFuncClusterID,
	); err != nil {
		setupLog.Error(err, "unable to create index for field", "field", "spec.buyerClusterID")
		os.Exit(1)
	}

	if err := cache.IndexField(
		context.Background(),
		&reservationv1alpha1.Contract{},
		"spec.buyer.additionalInformation.liqoID",
		indexFuncBuyerLiqoID,
	); err != nil {
		setupLog.Error(err, "unable to create index for field", "field", "spec.buyer.additionalInformation.liqoID")
		os.Exit(1)
	}

	if err := cache.IndexField(
		context.Background(),
		&reservationv1alpha1.Contract{},
		"spec.seller.additionalInformation.liqoID",
		indexFuncSellerLiqoID,
	); err != nil {
		setupLog.Error(err, "unable to create index for field", "field", "spec.seller.additionalInformation.liqoID")
		os.Exit(1)
	}

	gw := gateway.NewGateway(mgr.GetClient())
	grpcServer := grpc.NewGrpcServer(mgr.GetClient())

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

	if *enableWH {
		// Register Reservation webhook
		setupLog.Info("Registering webhooks to the manager")
		if err = (&reservationv1alpha1.Reservation{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Flavor")
			os.Exit(1)
		}
		if err = (&reservationv1alpha1.Contract{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Contract")
			os.Exit(1)
		}
		if err = (&reservationv1alpha1.Transaction{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Transaction")
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

	// Periodically clear the transaction cache
	if err := mgr.Add(manager.RunnableFunc(gw.CacheRefresher(flags.RefreshCacheInterval))); err != nil {
		klog.Errorf("Unable to set up transaction cache refresher: %s", err)
		os.Exit(1)
	}

	// Periodically check if Liqo is ready
	if err := mgr.Add(manager.RunnableFunc(gw.LiqoChecker(flags.LiqoCheckInterval))); err != nil {
		klog.Errorf("Unable to set up Liqo checker: %s", err)
		os.Exit(1)
	}

	// Start the REAR Gateway HTTP server
	if err := mgr.Add(manager.RunnableFunc(gw.Start)); err != nil {
		klog.Errorf("Unable to set up Gateway HTTP server: %s", err)
		os.Exit(1)
	}

	// Start the REAR GRPC server
	if err := mgr.Add(manager.RunnableFunc(grpcServer.Start)); err != nil {
		klog.Errorf("Unable to set up Gateway GRPC server: %s", err)
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
