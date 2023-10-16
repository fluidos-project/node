#!/usr/bin/bash

set -xeu

lrmImage="cannarelladev/local-resource-manager:v0.1"
rearManagerImage="cannarelladev/rear-manager:v0.1"
rearControllerImage="cannarelladev/rear-controller:v0.1"

consumer_node_port=30000
provider_node_port=30001

kind create cluster --config consumer/cluster-multi-worker.yaml  --name fluidos-consumer --kubeconfig $PWD/consumer/config
kind create cluster --config provider/cluster-multi-worker.yaml  --name fluidos-provider --kubeconfig $PWD/provider/config

consumer_controlplane_ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' fluidos-consumer-control-plane)
provider_controlplane_ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' fluidos-provider-control-plane)

export KUBECONFIG=$PWD/consumer/config

kubectl apply -f ../../deployments/node/crds
kubectl apply -f $PWD/metrics-server.yaml

echo "Waiting for metrics-server to be ready"
kubectl wait --for=condition=ready pod -l k8s-app=metrics-server -n kube-system --timeout=300s

kind load docker-image $lrmImage --name=fluidos-consumer
kind load docker-image $rearManagerImage --name=fluidos-consumer 
kind load docker-image $rearControllerImage --name=fluidos-consumer 

helm install node ../../deployments/node -n fluidos \
  --create-namespace -f consumer/values.yaml \
  --set networkManager.configMaps.nodeIdentity.ip=$consumer_controlplane_ip:$consumer_node_port \
  --set networkManager.configMaps.providers.local=$provider_controlplane_ip:$provider_node_port     

liqoctl install kind --cluster-name fluidos-consumer

export KUBECONFIG=$PWD/provider/config

kubectl apply -f ../../deployments/node/crds
kubectl apply -f $PWD/metrics-server.yaml

echo "Waiting for metrics-server to be ready"
kubectl wait --for=condition=ready pod -l k8s-app=metrics-server -n kube-system --timeout=300s

kind load docker-image $lrmImage --name=fluidos-provider
kind load docker-image $rearManagerImage --name=fluidos-provider
kind load docker-image $rearControllerImage --name=fluidos-provider

helm install node ../../deployments/node -n fluidos \
  --create-namespace -f provider/values.yaml \
  --set networkManager.configMaps.nodeIdentity.ip=$provider_controlplane_ip:$provider_node_port \
  --set networkManager.configMaps.providers.local=$consumer_controlplane_ip:$consumer_node_port

liqoctl install kind --cluster-name fluidos-provider


