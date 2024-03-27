// Copyright 2022-2023 FLUIDOS Project
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
	"net/http"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
	var probeAddr string
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&flags.AMOUNT, "amount", "", "Amount of money set for the flavours of this node")
	flag.StringVar(&flags.CURRENCY, "currency", "", "Currency of the money set for the flavours of this node")
	flag.StringVar(&flags.PERIOD, "period", "", "Period set for the flavours of this node")
	flag.StringVar(&flags.ResourceType, "resources-types", "k8s-fluidos", "Type of the Flavour related to k8s resources")
	flag.StringVar(&flags.CPUMin, "cpu-min", "0", "Minimum CPU value")
	flag.StringVar(&flags.MemoryMin, "memory-min", "0", "Minimum memory value")
	flag.StringVar(&flags.PodsMin, "pods-min", "0", "Minimum Pods value")
	flag.StringVar(&flags.CPUStep, "cpu-step", "0", "CPU step value")
	flag.StringVar(&flags.MemoryStep, "memory-step", "0", "Memory step value")
	flag.StringVar(&flags.PodsStep, "pods-step", "0", "Pods step value")
	flag.Int64Var(&flags.MinCount, "min-count", 0, "Minimum number of flavours")
	flag.Int64Var(&flags.MaxCount, "max-count", 0, "Maximum number of flavours")
	flag.StringVar(&flags.ResourceNodeLabel, "node-resource-label", "node-role.fluidos.eu/resources",
		"Label used to filter the k8s nodes from which create flavours")

	flag.Parse()

	cfg := ctrl.GetConfigOrDie()
	cl, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		setupLog.Error(err, "Unable to create client")
		os.Exit(1)
	}

	err = localresourcemanager.Start(context.Background(), cl)
	if err != nil {
		setupLog.Error(err, "Unable to start LocalResourceManager")
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthHandler) // health check endpoint
	mux.HandleFunc("/readyz", healthHandler)  // readiness check endpoint

	//nolint:gosec // We don't need this kind of security check
	server := &http.Server{
		Addr:    probeAddr,
		Handler: mux,
	}

	setupLog.Info("Starting server", "address", probeAddr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		setupLog.Error(err, "Server stopped unexpectedly")
		os.Exit(1)
	}
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}
