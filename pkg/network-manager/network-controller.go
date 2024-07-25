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

type ControlPlaneInfo struct {
	Address string `json:"address"`
}

type NetworkManager struct {
	discoveredControlPlanes map[string]ControlPlaneInfo
	mu                      sync.Mutex
	address                 string
}

func (nm *NetworkManager) sendMulticastMessage(multicastAddress string) error {
	info := ControlPlaneInfo{
		Address: nm.address,
	}
	message, err := json.Marshal(info)
	if err != nil {
		return err
	}

	conn, err := net.Dial("udp", multicastAddress)
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

	conn, err := net.ListenMulticastUDP("udp", nil, addr)
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

		fmt.Printf("Discovered control plane:  Address=%s\n", info.Address)
	}
}

func (nm *NetworkManager) printDiscoveredControlPlanes() {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	fmt.Println("Discovered Kubernetes Control Planes:")
	for _, cp := range nm.discoveredControlPlanes {
		fmt.Printf("Address: %s\n", cp.Address)
	}
}

func Start(ctx context.Context, cl client.Client) error {
	fmt.Println("Starting Kubernetes control plane discovery")

	multicastAddress := os.Getenv("MULTICAST_ADDRESS")
	if multicastAddress == "" {
		multicastAddress = "224.0.0.2:4000" // Default multicast address if not specified
	}

	clusterAddress, err := getClusterAddress()
	if err != nil {
		return fmt.Errorf("failed to get cluster address: %w", err)
	}

	nm := &NetworkManager{
		discoveredControlPlanes: make(map[string]ControlPlaneInfo),

		address: clusterAddress,
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

// These functions need to be implemented
func getControlPlaneID() (string, error) {
	id := os.Getenv("CONTROL_PLANE_ID")
	if id == "" {
		return "", fmt.Errorf("CONTROL_PLANE_ID environment variable not set")
	}
	return id, nil
}

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
