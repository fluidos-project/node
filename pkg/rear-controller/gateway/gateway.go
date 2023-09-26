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
//+kubebuilder:rbac:groups=reservation.github.com/fluidos-project/,resources=contracts,verbs=get;list;watch;create;update;patch;delete

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
	// mux creation
	router := mux.NewRouter()

	// routes definition
	router.HandleFunc("/api/listflavours", g.getFlavours).Methods("GET")
	router.HandleFunc("/api/listflavours/{flavourID}", g.getFlavourByID).Methods("GET")
	router.HandleFunc("/api/listflavours/selector", g.getFlavoursBySelector).Methods("POST")
	router.HandleFunc("/api/reserveflavour/{flavourID}", g.reserveFlavour).Methods("POST")
	router.HandleFunc("/api/purchaseflavour/{transactionID}", g.purchaseFlavour).Methods("POST")

	// Start server HTTP
	klog.Infof("Starting HTTP server on port %s", flags.HTTP_PORT)
	// TODO: after the demo recover correct address (14144)
	klog.Fatal(http.ListenAndServe(flags.HTTP_PORT, router))

}

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
