#!/usr/bin/bash

set -xeu

consumer_node_port=30000
provider_node_port=30001

kind create cluster --config consumer/cluster-multi-worker.yaml  --name fluidos-consumer --kubeconfig "$PWD/consumer/config"
kind create cluster --config provider/cluster-multi-worker.yaml  --name fluidos-provider --kubeconfig "$PWD/provider/config"

consumer_controlplane_ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' fluidos-consumer-control-plane)
provider_controlplane_ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' fluidos-provider-control-plane)

helm repo add fluidos https://fluidos-project.github.io/node/

export KUBECONFIG=$PWD/consumer/config

kubectl apply -f ../../deployments/node/crds
kubectl apply -f "$PWD/metrics-server.yaml"

echo "Waiting for metrics-server to be ready"
kubectl wait --for=condition=ready pod -l k8s-app=metrics-server -n kube-system --timeout=300s   

helm install node fluidos/node -n fluidos \
  --create-namespace -f consumer/values.yaml \
  --set networkManager.configMaps.nodeIdentity.ip="$consumer_controlplane_ip:$consumer_node_port" \
  --set networkManager.configMaps.providers.local="$provider_controlplane_ip:$provider_node_port"

liqoctl install kind --cluster-name fluidos-consumer \
  --set controllerManager.config.resourcePluginAddress=node-rear-controller-grpc.fluidos:2710 \
  --set controllerManager.config.enableResourceEnforcement=true

export KUBECONFIG=$PWD/provider/config

kubectl apply -f ../../deployments/node/crds
kubectl apply -f "$PWD/metrics-server.yaml"

echo "Waiting for metrics-server to be ready"
kubectl wait --for=condition=ready pod -l k8s-app=metrics-server -n kube-system --timeout=300s

helm install node fluidos/node -n fluidos \
  --create-namespace -f provider/values.yaml \
  --set networkManager.configMaps.nodeIdentity.ip="$provider_controlplane_ip:$provider_node_port" \
  --set networkManager.configMaps.providers.local="$consumer_controlplane_ip:$consumer_node_port"

liqoctl install kind --cluster-name fluidos-provider \
  --set controllerManager.config.resourcePluginAddress=node-rear-controller-grpc.fluidos:2710 \
  --set controllerManager.config.enableResourceEnforcement=true


