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

	"sigs.k8s.io/controller-runtime/pkg/client"

	clst "github.com/fluidos-project/node/apis/network/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/getters"
	"github.com/fluidos-project/node/pkg/utils/resourceforge"
)

// clusterRole
// +kubebuilder:rbac:groups=network.fluidos.eu,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=endpoints,verbs=get;list;watch

// ClusterInfo keeps the address of the discovered cluster.
type ClusterInfo struct {
	Address string `json:"address"`
}

// NetworkManager keeps all the necessary class data.
type NetworkManager struct {
	context            context.Context
	client             client.Client
	address            string
	iface              *net.Interface
	discoveredClusters map[string]*clst.Cluster
}

func (nm *NetworkManager) sendMulticastMessage(multicastAddress string) error {
	info := ClusterInfo{
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

		var info ClusterInfo

		err = json.Unmarshal(buffer[:n], &info)
		if err != nil {
			fmt.Printf("Error unmarshalling message: %v\n", err)
			continue
		}

		if info.Address != nm.address {
			_, found := nm.discoveredClusters[info.Address]

			if !found {
				fmt.Printf("Discovered new cluster:  Address=%s\n", info.Address)

				// Create Cluster CR reference
				cluster := resourceforge.ForgeCluster(info.Address)
				if err := nm.client.Create(nm.context, cluster); err != nil {
					fmt.Printf("Error when creating Cluster: %s", err)
					return err
				}
				fmt.Printf("Cluster %s created\n", cluster.Name)

				// Add recently discovered Cluster *CR to the map
				nm.discoveredClusters[info.Address] = cluster
			}
		}
	}
}

// Start function is the entrypoint of the Network Manager.
func Start(ctx context.Context, cl client.Client) error {
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
		discoveredClusters: make(map[string]*clst.Cluster),
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
