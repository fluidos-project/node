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

package grpc

import (
	context "context"
	"fmt"
	"log"
	"net"
	"sync"

	resourcemonitors "github.com/liqotech/liqo/pkg/liqo-controller-manager/resource-request-controller/resource-monitors"
	grpc "google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fluidos-project/node/pkg/utils/flags"
)

// Server is the object that contains all the logical data stractures of the REAR gRPC Server.
type Server struct {
	Server *grpc.Server
	client client.Client
	resourcemonitors.ResourceReaderServer
	subscribers sync.Map
}

// NewGrpcServer creates a new gRPC server.
func NewGrpcServer(cl client.Client) *Server {
	return &Server{
		Server: grpc.NewServer(),
		client: cl,
	}
}

// Start starts the gRPC server.
func (s *Server) Start(ctx context.Context) error {
	_ = ctx
	// server setup. The stream is not initialized here because it needs a subscriber so
	// it will be initialized in the Subscribe method below

	// gRPC Configuration
	klog.Info("Configuring gRPC Server")
	grpcURL := ":" + flags.GRPCPort
	lis, err := net.Listen("tcp", grpcURL)
	if err != nil {
		klog.Infof("gRPC failed to listen: %v", err)
		return fmt.Errorf("gRPC failed to listen: %w", err)
	}

	// register this server using the register interface defined in liqo
	resourcemonitors.RegisterResourceReaderServer(s.Server, s)
	klog.Infof("gRPC Server Listening on %s", grpcURL)
	// gRPC Server start listener
	if err := s.Server.Serve(lis); err != nil {
		return fmt.Errorf("gRPC server failed to serve: %w", err)
	}

	return nil
}

// ReadResources is the method that returns the resources assigned to a specific ClusterID.
func (s *Server) ReadResources(ctx context.Context, req *resourcemonitors.ClusterIdentity) (*resourcemonitors.PoolResourceList, error) {
	_ = ctx
	klog.Infof("Reading resources for cluster %s", req.ClusterID)
	resources, err := s.GetOfferResourcesByClusterID(req.ClusterID)
	if err != nil {
		// TODO: maybe should be returned an empty resource list
		return nil, err
	}

	log.Printf("Retrieved resources for clusterID %s: %v", req.ClusterID, resources)

	resourceList := []*resourcemonitors.ResourceList{{Resources: resources}}
	response := resourcemonitors.PoolResourceList{ResourceLists: resourceList}

	return &response, nil
}

// Subscribe is the method that subscribes a the Liqo controller manager to the gRPC server.
func (s *Server) Subscribe(req *resourcemonitors.Empty, srv resourcemonitors.ResourceReader_SubscribeServer) error {
	klog.Info("Liqo controller manager subscribed to the gRPC server")

	// Store the stream. Using req as key since each request will have a different req object.
	s.subscribers.Store(req, srv)
	ctx := srv.Context()

	// This notification is useful since you can edit the resources declared in the deployment and apply it to the cluster when one or more
	// foreign clusters are already peered so this broadcast notification will update the resources for those clusters.
	err := s.NotifyChange(context.Background(), &resourcemonitors.ClusterIdentity{ClusterID: resourcemonitors.AllClusterIDs})
	if err != nil {
		klog.Infof("Error during sending notification to liqo: %s", err)
	}

	for {
		<-ctx.Done()
		s.subscribers.Delete(req)
		klog.Infof("Liqo controller manager disconnected")
		return nil
	}
}

// NotifyChange is the method that notifies a change to the Liqo controller manager.
func (s *Server) NotifyChange(ctx context.Context, req *resourcemonitors.ClusterIdentity) error {
	_ = ctx
	klog.Infof("Notifying change to Liqo controller manager for cluster %s", req.ClusterID)
	var err error
	s.subscribers.Range(func(key, value interface{}) bool {
		stream := value.(resourcemonitors.ResourceReader_SubscribeServer)

		klog.Infof("Key: %v, Value: %v", key, value)

		err = stream.Send(req)
		if err != nil {
			err = fmt.Errorf("error: error during sending a notification %w", err)
		}
		return true
	})
	if err != nil {
		klog.Infof("%s", err)
		return err
	}
	klog.Infof("Notification sent to Liqo controller manager for cluster %s", req.ClusterID)
	return nil
}

// RemoveCluster is the method that removes a cluster from the gRPC server.
func (s *Server) RemoveCluster(ctx context.Context, req *resourcemonitors.ClusterIdentity) (*resourcemonitors.Empty, error) {
	_ = ctx
	klog.Infof("Removing cluster %s", req.ClusterID)
	klog.Info("Method RemoveCluster not implemented yet")
	// Implement here your logic
	return &resourcemonitors.Empty{}, nil
}

// GetOfferResourcesByClusterID is the method that returns the resources assigned to a specific ClusterID.
func (s *Server) GetOfferResourcesByClusterID(clusterID string) (map[string]*resource.Quantity, error) {
	log.Printf("Getting resources for cluster ID: %s", clusterID)
	resources, err := getContractResourcesByClusterID(s.client, clusterID)
	if err != nil {
		return nil, err
	}
	return resources, nil
}

// UpdatePeeringOffer is the method that updates the peering offer.
func (s *Server) UpdatePeeringOffer(clusterID string) {
	_ = s.NotifyChange(context.Background(), &resourcemonitors.ClusterIdentity{ClusterID: clusterID})
}
