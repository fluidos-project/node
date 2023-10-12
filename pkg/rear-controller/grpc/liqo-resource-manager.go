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

package grpc

import (
	context "context"
	"fmt"
	"log"
	"net"

	grpc "google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	resourcemonitors "github.com/liqotech/liqo/pkg/liqo-controller-manager/resource-request-controller/resource-monitors"

	"github.com/fluidos-project/node/pkg/utils/flags"
)

type grpcServer struct {
	Server *grpc.Server
	client client.Client
	//contractHandler connector.ContractHandler
	stream resourcemonitors.ResourceReader_SubscribeServer
	resourcemonitors.ResourceReaderServer
}

func NewGrpcServer(cl client.Client) *grpcServer {
	return &grpcServer{
		Server: grpc.NewServer(),
		client: cl,
	}
}

func (s *grpcServer) Start() {
	grpcUrl := ":" + flags.GRPC_PORT

	// gRPC Configuration
	klog.Info("Configuring gRPC Server")
	lis, err := net.Listen("tcp", grpcUrl)
	if err != nil {
		log.Fatalf("gRPC failed to listen: %v", err)
	}

	klog.Infof("gRPC Server Listening on %s", grpcUrl)
	// gRPC Server start listener
	if err := s.Server.Serve(lis); err != nil {
		log.Fatalf("gRPC failed to serve: %v", err)
	}
}

/* func (s *grpcServer) RegisterContractHandler(ch connector.ContractHandler) {
	s.contractHandler = ch
} */

func (s *grpcServer) ReadResources(ctx context.Context, req *resourcemonitors.ClusterIdentity) (*resourcemonitors.ResourceList, error) {
	readResponse := &resourcemonitors.ResourceList{Resources: map[string]*resource.Quantity{}}

	log.Printf("ReadResource for clusterID %s", req.ClusterID)
	resources, err := s.GetOfferResourcesByClusterID(req.ClusterID)
	if err != nil {
		// TODO: maybe should be resurned an empty resource list
		return nil, err
	}

	log.Printf("Retrieved resources for clusterID %s: %v", req.ClusterID, resources)
	for key, value := range *resources {
		readResponse.Resources[key.String()] = &value
	}

	return readResponse, nil
}

func (s *grpcServer) Subscribe(req *resourcemonitors.Empty, srv resourcemonitors.ResourceReader_SubscribeServer) error {
	// Implement here your logic
	s.stream = srv
	ctx := srv.Context()

	fmt.Println("Subscribe")

	s.NotifyChange(context.Background(), &resourcemonitors.ClusterIdentity{ClusterID: resourcemonitors.AllClusterIDs})

	for {
		select {
		case <-ctx.Done():
			klog.Info("liqo controller manager disconnected")
			return nil
		}
	}
}

func (s *grpcServer) NotifyChange(ctx context.Context, req *resourcemonitors.ClusterIdentity) error {
	// Implement here your logic
	if s.stream == nil {
		return fmt.Errorf("you must first subscribe a controller manager to notify a change")
	} else {
		s.stream.Send(req)
	}
	return nil
}

func (s *grpcServer) RemoveCluster(ctx context.Context, req *resourcemonitors.ClusterIdentity) (*resourcemonitors.Empty, error) {
	// Implement here your logic
	return nil, fmt.Errorf("Not implemented")
}

func (s *grpcServer) GetOfferResourcesByClusterID(clusterID string) (*corev1.ResourceList, error) {
	log.Printf("Getting resources for cluster ID: %s", clusterID)
	resources, err := getContractResourcesByClusterID(s.client, clusterID)
	if err != nil {
		return nil, err
	}
	return resources, nil
}

func (s *grpcServer) UpdatePeeringOffer(clusterID string) {
	s.NotifyChange(context.Background(), &resourcemonitors.ClusterIdentity{ClusterID: clusterID})
}
