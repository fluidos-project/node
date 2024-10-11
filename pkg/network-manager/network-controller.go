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
	"sync"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// clusterRole
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch

type ControlPlaneInfo struct {
	Address  string `json:"address"`
	Hostname string `json:"hostname"`
}

type NetworkManager struct {
	discoveredControlPlanes map[string]ControlPlaneInfo
	mu                      sync.Mutex
	address                 string
	hostname                string
	iface                   *net.Interface
}

func (nm *NetworkManager) sendMulticastMessage(multicastAddress string) error {
	info := ControlPlaneInfo{
		Address:  nm.address,
		Hostname: nm.hostname,
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

		var info ControlPlaneInfo
		err = json.Unmarshal(buffer[:n], &info)
		if err != nil {
			fmt.Printf("Error unmarshalling message: %v\n", err)
			continue
		}

		fmt.Printf("Discovered control plane:  Address=%s, Hostname=%s\n", info.Address, info.Hostname)
	}
}

func (nm *NetworkManager) printDiscoveredControlPlanes() {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	fmt.Println("Discovered Kubernetes Control Planes:")
	for _, cp := range nm.discoveredControlPlanes {
		fmt.Printf("Address: %s, Hostname: %s\n", cp.Address, cp.Hostname)
	}
}

func Start(ctx context.Context, cl client.Client) error {
	fmt.Println("Starting Kubernetes control plane discovery")

	multicastAddress := os.Getenv("MULTICAST_ADDRESS")
	if multicastAddress == "" {
		multicastAddress = "224.0.0.155:4000" // Default multicast address if not specified
	}

	clusterAddress, err := getClusterAddress()

	if err != nil {
		return fmt.Errorf("failed to get cluster address: %w", err)
	}

	clusterHostname, err := getClusterHostname()

	if err != nil {
		return fmt.Errorf("failed to get cluster hostname: %w", err)
	}

	ifi, err := net.InterfaceByName("eth1")
	if err != nil {
		return err
	}

	fmt.Printf("Interface: %s - MAC address: %s\n", ifi.Name, ifi.HardwareAddr)

	nm := &NetworkManager{
		discoveredControlPlanes: make(map[string]ControlPlaneInfo),

		address:  clusterAddress,
		hostname: clusterHostname,
		iface:    ifi,
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

	// Periodically print discovered control planes
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-ticker.C:
				nm.printDiscoveredControlPlanes()
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()

	// Block until context is cancelled
	<-ctx.Done()
	return nil
}

// // These functions need to be implemented
// func getControlPlaneID() (string, error) {
// 	id := os.Getenv("CONTROL_PLANE_ID")
// 	if id == "" {
// 		return "", fmt.Errorf("CONTROL_PLANE_ID environment variable not set")
// 	}
// 	return id, nil
// }

func getClusterAddress() (string, error) {
	// First, try to get the IP from the API server
	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	if host != "" {
		return host, nil
	}

	// If that fails, try to get the node's IP
	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("failed to get hostname: %v", err)
	}

	addrs, err := net.LookupIP(hostname)
	if err != nil {
		return "", fmt.Errorf("failed to lookup IPs: %v", err)
	}

	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil && !ipv4.IsLoopback() {
			return ipv4.String(), nil
		}
	}

	return "", fmt.Errorf("no suitable IPv4 address found for control plane")
}

func getClusterHostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("failed to get hostname: %v", err)
	}
	return hostname, nil
}
