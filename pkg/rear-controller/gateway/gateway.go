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

package gateway

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/consts"
	"github.com/fluidos-project/node/pkg/utils/flags"
	"github.com/fluidos-project/node/pkg/utils/getters"
	"github.com/fluidos-project/node/pkg/utils/models"
	"github.com/fluidos-project/node/pkg/utils/tools"
)

// clusterRole
//	+kubebuilder:rbac:groups=reservation.fluidos.eu,resources=contracts,verbs=get;list;watch;create;update;patch;delete
//	+kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=flavors,verbs=get;list;watch;create;update;patch;delete
//	+kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=flavors/status,verbs=get;update;patch
//	+kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=flavors/finalizers,verbs=update
// +kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=allocations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=allocations/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=allocations/finalizers,verbs=update
//	+kubebuilder:rbac:groups=core,resources=*,verbs=get;list;watch

const (
	// ListFlavorsPath is the path to get the list of flavors.
	ListFlavorsPath = "/api/listflavors"
	// ListFlavorByIDPath is the path to get a flavor by ID.
	ListFlavorByIDPath = "/api/listflavors/"
	// ReserveFlavorPath is the path to reserve a flavor.
	ReserveFlavorPath = "/api/reserveflavor/"
	// PurchaseFlavorPath is the path to purchase a flavor.
	PurchaseFlavorPath = "/api/purchaseflavor/"
	// ListFlavorsBySelectorPath is the path to get the list of flavors by selector.
	ListFlavorsBySelectorPath = "/api/listflavors/selector"
)

// Gateway is the object that contains all the logical data stractures of the REAR Gateway.
type Gateway struct {
	// NodeIdentity is the identity of the FLUIDOS Node
	ID *nodecorev1alpha1.NodeIdentity

	// Transactions is a map of Transaction
	Transactions map[string]*models.Transaction

	// client is the Kubernetes client
	client client.Client

	// Readyness of the Gateway. It is set when liqo is installed
	LiqoReady bool

	// The Liqo ClusterID
	ClusterID string
}

// NewGateway creates a new Gateway object.
func NewGateway(c client.Client) *Gateway {
	return &Gateway{
		client:       c,
		Transactions: make(map[string]*models.Transaction),
		LiqoReady:    false,
		ClusterID:    "",
	}
}

// Start starts a new HTTP server.
func (g *Gateway) Start(ctx context.Context) error {
	klog.Info("Getting FLUIDOS Node identity...")

	nodeIdentity := getters.GetNodeIdentity(ctx, g.client)
	if nodeIdentity == nil {
		klog.Info("Error getting FLUIDOS Node identity")
		return fmt.Errorf("error getting FLUIDOS Node identity")
	}

	g.RegisterNodeIdentity(nodeIdentity)

	router := mux.NewRouter()

	//nolint:gocritic // middleware for debugging purposes
	// router.Use(loggingMiddleware)

	// middleware for readiness
	router.Use(g.readinessMiddleware)

	// Gateway endpoints
	router.HandleFunc(Routes.Flavors, g.getFlavors).Methods("GET")
	router.HandleFunc(Routes.K8SliceFlavors, g.getK8SliceFlavorsBySelector).Methods("GET")
	// TODO: Implement the other selector types
	router.HandleFunc(Routes.Reserve, g.reserveFlavor).Methods("POST")
	router.HandleFunc(Routes.Purchase, g.purchaseFlavor).Methods("POST")

	// Configure the HTTP server
	//nolint:gosec // we are not using a TLS certificate
	srv := &http.Server{
		Handler: router,
		Addr:    ":" + flags.HTTPPort,
	}

	// Start server HTTP
	klog.Infof("Starting HTTP server on port %s", flags.HTTPPort)
	return srv.ListenAndServe()
}

// RegisterNodeIdentity registers the FLUIDOS Node identity into the Gateway.
func (g *Gateway) RegisterNodeIdentity(nodeIdentity *nodecorev1alpha1.NodeIdentity) {
	g.ID = nodeIdentity
}

func (g *Gateway) readinessMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !g.LiqoReady {
			klog.Infof("Liqo not ready yet")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// CacheRefresher is a function that periodically checks the cache and removes expired transactions.
func (g *Gateway) CacheRefresher(interval time.Duration) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		return wait.PollUntilContextCancel(ctx, interval, false, g.refreshCache)
	}
}

// check expired transactions and remove them from the cache.
//
//nolint:revive // we need to pass ctx as parameter to be compliant with the Poller interface
func (g *Gateway) refreshCache(ctx context.Context) (bool, error) {
	klog.InfoDepth(1, "Refreshing cache...")
	for transactionID, transaction := range g.Transactions {
		if tools.CheckExpiration(transaction.ExpirationTime) {
			klog.Infof("Transaction %s expired, removing it from cache...", transactionID)
			g.removeTransaction(transactionID)
			return false, nil
		}
	}
	return false, nil
}

// LiqoChecker is a function that periodically checks if Liqo is ready.
func (g *Gateway) LiqoChecker(interval time.Duration) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		return wait.PollUntilContextCancel(ctx, interval, false, g.checkLiqoReadiness)
	}
}

// check if Liqo is ready and set the LiqoReady flag to true.
func (g *Gateway) checkLiqoReadiness(ctx context.Context) (bool, error) {
	klog.Infof("Checking Liqo readiness")
	if g.LiqoReady && g.ClusterID != "" {
		return true, nil
	}

	var cm corev1.ConfigMap
	err := g.client.Get(ctx, types.NamespacedName{Name: consts.LiqoClusterIdConfigMapName, Namespace: consts.LiqoNamespace}, &cm)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			klog.Errorf("Error when retrieving Liqo ConfigMap: %s", err)
		}
		klog.Infof("Liqo not ready yet. ConfigMap not found")
		return false, nil
	}

	if cm.Data["CLUSTER_ID"] != "" && cm.Data["CLUSTER_NAME"] != "" {
		klog.Infof("Liqo is ready")
		g.LiqoReady = true
		g.ClusterID = cm.Data["CLUSTER_ID"]
		return true, nil
	}

	klog.Infof("Liqo not ready yet")
	return false, nil
}
