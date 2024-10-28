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
	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/flags"
	"github.com/fluidos-project/node/pkg/utils/getters"
	"github.com/fluidos-project/node/pkg/utils/namings"
	"github.com/fluidos-project/node/pkg/utils/resourceforge"
	"github.com/fluidos-project/node/pkg/utils/tools"
)

// clusterRole
// +kubebuilder:rbac:groups=network.fluidos.eu,resources=knownclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=network.fluidos.eu,resources=knownclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=endpoints,verbs=get;list;watch

// NetworkManager keeps all the necessary class data.
type NetworkManager struct {
	ID                   *nodecorev1alpha1.NodeIdentity
	Multicast            string
	Iface                *net.Interface
	EnableLocalDiscovery bool
}

// KnownClusterReconciler reconciles a KnownCluster object.
type KnownClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile reconciles a KnownClusters from DiscoveredClustersList.
func (r *KnownClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx, "kowncluster", req.NamespacedName)
	ctx = ctrl.LoggerInto(ctx, log)

	klog.InfoS("Reconcile triggered", "context", ctx)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KnownClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkv1alpha1.KnownCluster{}).
		Complete(r)
}

// Setup the Network Manager.
func Setup(ctx context.Context, cl client.Client, nm *NetworkManager, cniInterface *string) error {
	klog.Info("Setting up Network Manager routines")

	nodeIdentity := getters.GetNodeIdentity(ctx, cl)
	if nodeIdentity == nil {
		return fmt.Errorf("failed to get Node Identity")
	}

	multicastAddress := os.Getenv("MULTICAST_ADDRESS")
	if multicastAddress == "" {
		return fmt.Errorf("failed to get multicast address")
	}

	nm.ID = nodeIdentity
	nm.Multicast = multicastAddress

	if nm.EnableLocalDiscovery {
		ifi, err := net.InterfaceByName(*cniInterface)
		if err != nil {
			return err
		}
		nm.Iface = ifi
		klog.InfoS("Interface", "Name", ifi.Name, "MAC address", ifi.HardwareAddr)
	}

	klog.InfoS("Node", "ID", nodeIdentity.NodeID, "Address", nodeIdentity.IP)

	return nil
}

// Execute the Network Manager routines.
func Execute(ctx context.Context, cl client.Client, nm *NetworkManager) error {
	// Start sending multicast messages
	if nm.EnableLocalDiscovery {
		go func() {
			if err := sendMulticastMessage(ctx, nm); err != nil {
				klog.ErrorS(err, "Error sending advertisemente")
			}
		}()

		// Start receiving multicast messages
		go func() {
			if err := receiveMulticastMessage(ctx, cl, nm); err != nil {
				klog.ErrorS(err, "Error receiving advertisement")
			}
		}()
	}

	// Do housekeeping
	go func() {
		if err := doHousekeeping(ctx, cl); err != nil {
			klog.ErrorS(err, "Error doing housekeeping")
		}
	}()

	return nil
}

func sendMulticastMessage(ctx context.Context, nm *NetworkManager) error {
	message, err := json.Marshal(nm.ID)
	if err != nil {
		return err
	}

	laddr, err := nm.Iface.Addrs()
	if err != nil {
		return err
	}

	dialer := &net.Dialer{
		LocalAddr: &net.UDPAddr{
			IP:   laddr[0].(*net.IPNet).IP,
			Port: 0,
		},
	}

	conn, err := dialer.Dial("udp", nm.Multicast)
	if err != nil {
		return err
	}
	defer conn.Close()

	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			_, err = conn.Write(message)
			if err != nil {
				return err
			}
			klog.Info("Advertisement multicasted")
		case <-ctx.Done():
			ticker.Stop()
			return nil
		}
	}
}

func receiveMulticastMessage(ctx context.Context, cl client.Client, local *NetworkManager) error {
	addr, err := net.ResolveUDPAddr("udp", local.Multicast)
	if err != nil {
		return err
	}

	conn, err := net.ListenMulticastUDP("udp", local.Iface, addr)
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

		var remote NetworkManager

		err = json.Unmarshal(buffer[:n], &remote.ID)
		if err != nil {
			klog.Error("Error unmarshalling message: ", err)
			continue
		}

		// Check if received advertisement is remote
		if local.ID.IP != remote.ID.IP {
			klog.InfoS("Received remote advertisement", "ID", remote.ID.NodeID, "Address", remote.ID.IP)

			// Fetch the KnownCluster instance if already present
			kc := &networkv1alpha1.KnownCluster{}

			if err := cl.Get(ctx, client.ObjectKey{Name: namings.ForgeKnownClusterName(remote.ID.NodeID), Namespace: flags.FluidosNamespace}, kc); err != nil {
				if client.IgnoreNotFound(err) == nil {
					klog.Info("KnownCluster not found: creating")

					// Create new KnownCluster CR
					if err := cl.Create(ctx, resourceforge.ForgeKnownCluster(remote.ID.NodeID, remote.ID.IP)); err != nil {
						return err
					}
					klog.InfoS("KnownCluster created", "ID", remote.ID.NodeID)
				}
			} else {
				klog.Info("KnownCluster already present: updating")
				// Update Status
				kc.UpdateStatus()

				// Update fetched KnownCluster CR
				err := cl.Status().Update(ctx, kc)
				if err != nil {
					return err
				}
				klog.InfoS("KnownCluster updated", "ID", kc.ObjectMeta.Name)
			}
		}
	}
}

func doHousekeeping(ctx context.Context, cl client.Client) error {
	ticker := time.NewTicker(20 * time.Second)
	for {
		select {
		case <-ticker.C:
			klog.Info("Starting housekeeping")

			// Retrieve KnownClusterList
			kcList := networkv1alpha1.KnownClusterList{}
			err := cl.List(ctx, &kcList)
			if err != nil {
				return err
			}
			if len(kcList.Items) == 0 {
				klog.Info("Housekeeping not needed, no KnownClusters available")
				continue
			}

			// Remove all KnownCluster CR with expiration time < now
			for i := range kcList.Items {
				kc := &kcList.Items[i]
				if tools.CheckExpiration(kc.Status.ExpirationTime) {
					err := cl.Delete(ctx, kc)
					klog.InfoS("KnownCluster expired and deleted", "ID", kc.Name)
					if err != nil {
						return err
					}
				}
			}
		case <-ctx.Done():
			ticker.Stop()
			return nil
		}
	}
}
