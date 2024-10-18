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

package networkmanager

import (
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	networkv1alpha1 "github.com/fluidos-project/node/apis/network/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/getters"
	"github.com/fluidos-project/node/pkg/utils/resourceforge"
)

// clusterRole
// +kubebuilder:rbac:groups=network.fluidos.eu,resources=knownclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=endpoints,verbs=get;list;watch

// KnownClusterReconciler reconciles a KnownCluster object.
type KnownClusterReconciler struct {
	client.Client
	Scheme                 *runtime.Scheme
	DiscoveredClustersList list.List
}

// Reconcile reconciles a KnownClusters from DiscoveredClustersList.
func (r *KnownClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx, "knowncluster", req.NamespacedName)
	ctx = ctrl.LoggerInto(ctx, log)

	klog.Info("Reconcile started")

	e := r.DiscoveredClustersList.Front() // FIFO
	if e == nil {
		return ctrl.Result{}, nil
	}

	id := e.Value.(KnownClusterInfo).ID
	address := e.Value.(KnownClusterInfo).Address

	// Fetch the KnownCluster instance
	var clst networkv1alpha1.KnownCluster
	if err := r.Get(ctx, client.ObjectKey{Name: id}, &clst); err != nil {
		if client.IgnoreNotFound(err) != nil {
			klog.Info("KnownCluster not found: creating")

			// Create Cluster CR reference
			cluster := resourceforge.ForgeKnownCluster(id, address)
			if err := r.Client.Create(ctx, cluster); err != nil {
				fmt.Printf("Error when creating Cluster: %s", err)
				return ctrl.Result{}, err
			}
			fmt.Printf("Cluster %s created\n", cluster.Name)
			r.DiscoveredClustersList.Remove(e)
		}
	} else {
		// Update var
		clst.Status.LastUpdateTime = time.Now().Local().UnixMilli()
		clst.Status.ExpirationTime = time.Now().Local().Add(time.Second * 10).UnixMilli()

		// Update CR
		err := r.Client.Status().Update(ctx, &clst)
		if err != nil {
			return ctrl.Result{}, err
		}
		r.DiscoveredClustersList.Remove(e)
	}
	// Remove all the timed-out CRs
	err := houseKeeping(ctx, r)

	return ctrl.Result{}, err
}

// ClusterInfo keeps the address of the discovered cluster.
type KnownClusterInfo struct {
	ID      string `json:"id"`
	Address string `json:"address"`
}

// NetworkManager keeps all the necessary class data.
type NetworkManager struct {
	context            context.Context
	client             client.Client
	id                 string
	address            string
	iface              *net.Interface
	discoveredClusters *list.List
}

func (nm *NetworkManager) sendMulticastMessage(multicastAddress string) error {
	info := KnownClusterInfo{
		ID:      nm.id,
		Address: nm.address,
	}

	message, err := json.Marshal(info)
	if err != nil {
		return err
	}

	laddr, err := nm.iface.Addrs()
	if err != nil {
		return err
	}

	dialer := &net.Dialer{
		LocalAddr: &net.UDPAddr{
			IP:   laddr[0].(*net.IPNet).IP,
			Port: 0,
		},
	}

	conn, err := dialer.Dial("udp", multicastAddress)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(message)
	return err
}

func (nm *NetworkManager) receiveMulticastMessage(multicastAddress string) error {
	addr, err := net.ResolveUDPAddr("udp", multicastAddress)
	if err != nil {
		return err
	}

	conn, err := net.ListenMulticastUDP("udp", nm.iface, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	buffer := make([]byte, 1024)

	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			return err
		}

		var info KnownClusterInfo

		err = json.Unmarshal(buffer[:n], &info)
		if err != nil {
			fmt.Printf("Error unmarshalling message: %v\n", err)
			continue
		}

		// Enqueue any announcing cluster
		if nm.address != info.Address {
			fmt.Printf("Discovered new cluster:  ID=%s - Address=%s\n", info.ID, info.Address)
			nm.discoveredClusters.PushBack(info)
		}
	}
}

// Start function is the entrypoint of the Network Manager.
func StartDiscovery(ctx context.Context, cl client.Client, discClusters *list.List) error {
	fmt.Println("Starting Kubernetes cluster discovery")

	multicastAddress := os.Getenv("MULTICAST_ADDRESS")
	if multicastAddress == "" {
		return fmt.Errorf("failed to get multicast address")
	}

	clusterAddress, err := getClusterAddress(ctx, cl)
	if err != nil {
		return fmt.Errorf("failed to get cluster address: %w", err)
	}

	ifi, err := net.InterfaceByName("eth1")
	if err != nil {
		return err
	}

	fmt.Printf("Interface: %s - MAC address: %s - Address: %s\n", ifi.Name, ifi.HardwareAddr, clusterAddress)

	nm := &NetworkManager{
		context:            ctx,
		client:             cl,
		discoveredClusters: discClusters,
		address:            clusterAddress,
		iface:              ifi,
	}

	// Start sending multicast messages
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-ticker.C:
				if err := nm.sendMulticastMessage(multicastAddress); err != nil {
					fmt.Printf("Error sending multicast message: %v\n", err)
				}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()

	// Start receiving multicast messages
	go func() {
		if err := nm.receiveMulticastMessage(multicastAddress); err != nil {
			fmt.Printf("Error receiving multicast message: %v\n", err)
		}
	}()

	// Block until context is canceled
	<-ctx.Done()
	return nil
}

func getClusterAddress(ctx context.Context, cl client.Client) (string, error) {
	// Get NodeIdentity
	nodeIdentity := getters.GetNodeIdentity(ctx, cl)
	if nodeIdentity != nil {
		return nodeIdentity.IP, nil
	}

	return "", fmt.Errorf("no suitable IPv4 address found for cluster")
}

// SetupWithManager sets up the controller with the Manager.
func (r *KnownClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkv1alpha1.KnownCluster{}).
		Complete(r)
}

// Delete all the expired KnownCluster CRs.
func houseKeeping(ctx context.Context, r *KnownClusterReconciler) error {
	// Retrieve the CR list
	clstList := networkv1alpha1.KnownClusterList{}
	err := r.Client.List(ctx, &clstList)
	if err != nil {
		return err
	}
	// Remove all CR with expiration time < now
	for _, cr := range clstList.Items {
		if cr.Status.ExpirationTime < time.Now().Local().UnixMilli() {
			err := r.Client.Delete(ctx, &cr)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
