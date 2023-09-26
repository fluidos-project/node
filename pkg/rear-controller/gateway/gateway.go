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

package gateway

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/flags"
	"github.com/fluidos-project/node/pkg/utils/models"
	"github.com/fluidos-project/node/pkg/utils/tools"
)

// clusterRole
//+kubebuilder:rbac:groups=reservation.fluidos.eu,resources=contracts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=flavours,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=flavours/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=flavours/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch

const (
	LIST_FLAVOURS_PATH             = "/api/listflavours"
	LIST_FLAVOUR_BY_ID_PATH        = "/api/listflavours/"
	RESERVE_FLAVOUR_PATH           = "/api/reserveflavour/"
	PURCHASE_FLAVOUR_PATH          = "/api/purchaseflavour/"
	LIST_FLAVOURS_BY_SELECTOR_PATH = "/api/listflavours/selector"
)

type Gateway struct {
	// NodeIdentity is the identity of the FLUIDOS Node
	ID *nodecorev1alpha1.NodeIdentity

	// Transactions is a map of Transaction
	Transactions map[string]models.Transaction

	// client is the Kubernetes client
	client client.Client
}

func NewGateway(c client.Client) *Gateway {
	return &Gateway{
		client:       c,
		Transactions: make(map[string]models.Transaction),
	}
}

// StartHttpServer starts a new HTTP server
func (g *Gateway) StartHttpServer() {
	router := mux.NewRouter()

	// middleware for debugging purposes
	// router.Use(loggingMiddleware)

	// Gateway endpoints
	router.HandleFunc(LIST_FLAVOURS_PATH, g.getFlavours).Methods("GET")
	router.HandleFunc(LIST_FLAVOUR_BY_ID_PATH+"{flavourID}", g.getFlavourByID).Methods("GET")
	router.HandleFunc(LIST_FLAVOURS_BY_SELECTOR_PATH, g.getFlavoursBySelector).Methods("POST")
	router.HandleFunc(RESERVE_FLAVOUR_PATH+"{flavourID}", g.reserveFlavour).Methods("POST")
	router.HandleFunc(PURCHASE_FLAVOUR_PATH+"{transactionID}", g.purchaseFlavour).Methods("POST")

	// Configure the HTTP server
	srv := &http.Server{
		Handler: router,
		Addr:    ":" + flags.HTTP_PORT,
	}

	// Start server HTTP
	klog.Infof("Starting HTTP server on port %s", flags.HTTP_PORT)
	klog.Fatal(srv.ListenAndServe())

}

// Only for debugging purposes
/* func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		klog.Infof("Received request: %s %s %v", r.Method, r.URL.Path, r)
		next.ServeHTTP(w, r)
	})
} */

func (g *Gateway) CacheRefresher(interval time.Duration) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		return wait.PollUntilContextCancel(ctx, interval, false, g.refreshCache)
	}
}

// check expired transactions and remove them from the cache
func (g *Gateway) refreshCache(ctx context.Context) (bool, error) {
	klog.Infof("Refreshing cache")
	for transactionID, transaction := range g.Transactions {
		if tools.CheckExpiration(transaction.StartTime, flags.EXPIRATION_TRANSACTION) {
			klog.Infof("Transaction %s expired, removing it from cache...", transactionID)
			g.removeTransaction(transactionID)
			return false, nil
		}
	}
	return false, nil
}
