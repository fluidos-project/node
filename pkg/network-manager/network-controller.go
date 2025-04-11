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
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;delete

// NetworkManager keeps all the necessary class data.
type NetworkManager struct {
	ID                   *nodecorev1alpha1.NodeIdentity
	Multicast            string
	Iface                *net.Interface
	EnableLocalDiscovery bool
}

// BrokerReconciler reconciles a Broker object.
type BrokerReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	ActiveBrokers []*BrokerClient
}

// Reconcile reconciles a Broker.
func (r *BrokerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.InfoS("Reconcile triggered")
	log := ctrl.LoggerFrom(ctx, "broker", req.NamespacedName)
	ctx = ctrl.LoggerInto(ctx, log)

	var broker networkv1alpha1.Broker

	if err := r.Get(ctx, req.NamespacedName, &broker); client.IgnoreNotFound(err) != nil {
		klog.Errorf("Error when getting Broker %s before reconcile: %v", req.NamespacedName, err)
		return ctrl.Result{}, err
	} else if err != nil {
		klog.Infof("Broker %s not found, probably deleted", req.NamespacedName)

		// Deleting BrokerClient
		for i, brokerCl := range r.ActiveBrokers {
			if brokerCl.brokerName == req.NamespacedName.Name {
				klog.Infof("Deleting %s ", brokerCl.brokerName)
				err = r.brokerDelete(brokerCl, i)
				for i := range r.ActiveBrokers {
					klog.Info(r.ActiveBrokers[i])
				}
				if err != nil {
					return ctrl.Result{}, err
				}
			}
		}
		return ctrl.Result{}, nil
	}

	// If found in CR && found in slice -> update
	found := false
	sameAddr := false
	for i, brokerCl := range r.ActiveBrokers {
		if brokerCl.brokerName == broker.Spec.Name {
			klog.Info("found brokerClient ", brokerCl.brokerName)
			klog.Info("and Broker ", broker.Spec.Name)
			found = true

			// Update
			if err := r.brokerUpdate(&broker, brokerCl, i); err != nil {
				klog.Error("brokerUpdate failed: %s", err)
				if err = cleanBroker(ctx, r.Client, &broker); err != nil {
					klog.Error("cleanBroker failed: %s", err)
				}
				return ctrl.Result{}, err
			}
			break
		} else if brokerCl.serverAddr == broker.Spec.Address {
			sameAddr = true
			klog.Error("brokerUpdate failed: same address found in another BrokerClient")
			if err := cleanBroker(ctx, r.Client, &broker); err != nil {
				klog.Error("cleanBroker failed: %s", err)
			}
		}
	}
	// If found in CR && !found in slice && address is free -> create
	if (!found) && (!sameAddr) {
		// Create
		if err := r.brokerCreate(&broker); err != nil {
			klog.Error("brokerCreate failed: %s", err)
			if err = cleanBroker(ctx, r.Client, &broker); err != nil {
				klog.Error("cleanBroker failed: %s", err)
			}
			return ctrl.Result{}, err
		}
	}
	klog.Infof("Reconciling Broker %s", req.NamespacedName)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BrokerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkv1alpha1.Broker{}).
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

// Update the clientBroker.
func (r *BrokerReconciler) brokerUpdate(broker *networkv1alpha1.Broker, brokerCl *BrokerClient, index int) error {
	klog.Infof("updating broker: %s", brokerCl.brokerName)
	if err := r.brokerDelete(brokerCl, index); err != nil {
		return err
	}
	if err := r.brokerCreate(broker); err != nil {
		return err
	}
	return nil
}

// Create the clientBroker.
func (r *BrokerReconciler) brokerCreate(broker *networkv1alpha1.Broker) error {
	var bc BrokerClient
	var err error
	if err = bc.SetupBrokerClient(r.Client, broker); err != nil {
		return err
	}
	if err = bc.ExecuteBrokerClient(r.Client); err != nil {
		return err
	}
	r.ActiveBrokers = append(r.ActiveBrokers, &bc)
	return nil
}

// Delete the clientBroker.
func (r *BrokerReconciler) brokerDelete(brokerCl *BrokerClient, index int) error {
	if err := brokerCl.brokerConn.amqpChan.Close(); err != nil {
		klog.Errorf("Failed to close channel for broker '%s': %v", brokerCl.brokerName, err)
		return err
	}
	if err := brokerCl.brokerConn.amqpConn.Close(); err != nil {
		klog.Errorf("Failed to close connection for broker '%s': %v", brokerCl.brokerName, err)
		return err
	}
	brokerCl.canc()
	r.ActiveBrokers = append(r.ActiveBrokers[:index], r.ActiveBrokers[index+1:]...)
	return nil
}

// Delete the Broker CR.
func cleanBroker(ctx context.Context, cl client.Client, broker *networkv1alpha1.Broker) error {
	err := cl.Delete(ctx, broker)
	if err != nil {
		klog.Error("error during Broker deletion '%s': %w", broker.Spec.Name, err)
		return err
	}
	return nil
}
