#!/usr/bin/bash

read -p "Enable edge worker nodes support for the Fluidos provider cluster? [y/N]: " edge_ena
edge_ena=${edge_ena:-N}

if [ $edge_ena == "y" -o $edge_ena == "Y" ]; then
  edge_ena=1
else
  edge_ena=0
fi

if [ $edge_ena -eq 1 ]; then
  # Download and extract CNI plugins at ./cni/bin
  echo "Download and extract CNI plugins at ./cni/bin"
  VER=$(curl -s https://api.github.com/repos/containernetworking/plugins/releases | grep tag_name | cut -d '"' -f 4 | sed 1q | sed 's/v//g')
  wget -q https://github.com/containernetworking/plugins/releases/download/v$VER/cni-plugins-linux-amd64-v$VER.tgz
  mkdir -p ./cni/bin
  tar Cxzf ./cni/bin cni-plugins-linux-amd64-v$VER.tgz
  rm cni-plugins-linux-amd64-v$VER.tgz
  # Update kind configuration file with hostPath with CNI plugin binaries absolute path
  echo "Update kind configuration file with hostPath with CNI plugin binaries absolute path"
  sed -i 's, hostPath:.*, hostPath: '"$PWD"'/cni/bin,' provider/cluster-multi-worker-edge.yaml
fi

set -xeu

consumer_node_port=30000
provider_node_port=30001

kind create cluster --config consumer/cluster-multi-worker.yaml  --name fluidos-consumer --kubeconfig "$PWD/consumer/config"
if [ $edge_ena -eq 1 ]; then
  kind create cluster --config provider/cluster-multi-worker-edge.yaml  --name fluidos-provider --kubeconfig "$PWD/provider/config"
else
  kind create cluster --config provider/cluster-multi-worker.yaml  --name fluidos-provider --kubeconfig "$PWD/provider/config"
fi

consumer_controlplane_ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' fluidos-consumer-control-plane)
provider_controlplane_ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' fluidos-provider-control-plane)

helm repo add fluidos https://fluidos-project.github.io/node/

export KUBECONFIG=$PWD/consumer/config

kubectl apply -f ../../deployments/node/crds
kubectl apply -f "$PWD/metrics-server.yaml"

echo "Waiting for metrics-server to be ready"
# sleep to avoid error: no matching resources found
sleep 10
kubectl wait --for=condition=ready pod -l k8s-app=metrics-server -n kube-system --timeout=300s   

helm install node fluidos/node -n fluidos \
  --create-namespace -f consumer/values.yaml \
  --set networkManager.configMaps.nodeIdentity.ip="$consumer_controlplane_ip:$consumer_node_port" \
  --set networkManager.configMaps.providers.local="$provider_controlplane_ip:$provider_node_port"

liqoctl install kind --cluster-name fluidos-consumer \
  --set controllerManager.config.resourcePluginAddress=node-rear-controller-grpc.fluidos:2710 \
  --set controllerManager.config.enableResourceEnforcement=true

export KUBECONFIG=$PWD/provider/config

if [ $edge_ena -eq 1 ]; then

  kubectl taint nodes fluidos-provider-control-plane node-role.kubernetes.io/control-plane-
  kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml
  echo "Waiting for nodes to be ready"
  kubectl wait --for=condition=ready node -l kubernetes.io/hostname=fluidos-provider-control-plane --timeout=300s
  kubectl wait --for=condition=ready node -l kubernetes.io/hostname=fluidos-provider-worker --timeout=300s
  kubectl wait --for=condition=ready node -l kubernetes.io/hostname=fluidos-provider-worker2 --timeout=300s
  kubectl patch daemonset kube-proxy --context kind-fluidos-provider -n kube-system -p '{"spec": {"template": {"spec": {"affinity": {"nodeAffinity": {"requiredDuringSchedulingIgnoredDuringExecution": {"nodeSelectorTerms": [{"matchExpressions": [{"key": "node-role.kubernetes.io/edge", "operator": "DoesNotExist"}]}]}}}}}}}'
  kubectl patch deploy coredns --context kind-fluidos-provider -n kube-system -p '{"spec": {"template": {"spec": {"affinity": {"nodeAffinity": {"requiredDuringSchedulingIgnoredDuringExecution": {"nodeSelectorTerms": [{"matchExpressions": [{"key": "node-role.kubernetes.io/edge", "operator": "DoesNotExist"}]}]}}}}}}}'
  kubectl patch daemonset kube-flannel-ds --context kind-fluidos-provider -n kube-flannel -p '{"spec": {"template": {"spec": {"affinity": {"nodeAffinity": {"requiredDuringSchedulingIgnoredDuringExecution": {"nodeSelectorTerms": [{"matchExpressions": [{"key": "node-role.kubernetes.io/edge", "operator": "DoesNotExist"}]}]}}}}}}}'
  kubectl patch daemonset liqo-route --context kind-fluidos-provider -n liqo -p '{"spec": {"template": {"spec": {"affinity": {"nodeAffinity": {"requiredDuringSchedulingIgnoredDuringExecution": {"nodeSelectorTerms": [{"matchExpressions": [{"key": "node-role.kubernetes.io/edge", "operator": "DoesNotExist"}]}]}}}}}}}'

  rm -rf FluidosEdge-System
  git clone -n --depth=1 --filter=tree:0 https://github.com/otto-tom/FluidosEdge-System.git
  pushd FluidosEdge-System
  git sparse-checkout set --no-cone manifests/charts/cloudcore
  git checkout
  popd
  rm -rf manifests
  mv FluidosEdge-System/manifests/ ./
  rm -rf FluidosEdge-System

  DNS_NAME=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' fluidos-provider-control-plane)
  ADV_ADDR=$(ip -o route get to 8.8.8.8 | sed -n 's/.*src \([0-9.]\+\).*/\1/p')
  #read -p "Host IP is $ADV_ADDR. If this is correct press enter or provide it manualy: " UADV_ADDR
  #UADV_ADDR=${UADV_ADDR:-N}
  #if [ "$UADV_ADDR" != "N" ]; then
  #  ADV_ADDR=$UADV_ADDR
  #fi

  echo "Installing Cloudcore"
  helm upgrade --install cloudcore ./manifests/charts/cloudcore --namespace kubeedge --create-namespace -f ./manifests/charts/cloudcore/values.yaml --set cloudCore.modules.cloudHub.advertiseAddress[0]=$ADV_ADDR --set cloudCore.modules.cloudHub.dnsNames[0]=$DNS_NAME

  echo "Waiting for cloudcore to be ready"
  kubectl wait --for=condition=ready pod -l kubeedge=cloudcore -n kubeedge --timeout=300s
  set +xeu
  echo "Waiting for cloudcore tokensecret to be ready"
  while ! kubectl get secret tokensecret -nkubeedge; do sleep 2; done
  set -xeu
  KE_TOKEN=$(kubectl get secret -nkubeedge tokensecret -o=jsonpath='{.data.tokendata}' | base64 -d)
fi

kubectl apply -f ../../deployments/node/crds
kubectl apply -f "$PWD/metrics-server.yaml"

echo "Waiting for metrics-server to be ready"
# sleep to avoid error: no matching resources found
sleep 10
kubectl wait --for=condition=ready pod -l k8s-app=metrics-server -n kube-system --timeout=300s

helm install node fluidos/node -n fluidos \
  --create-namespace -f provider/values.yaml \
  --set networkManager.configMaps.nodeIdentity.ip="$provider_controlplane_ip:$provider_node_port" \
  --set networkManager.configMaps.providers.local="$consumer_controlplane_ip:$consumer_node_port"

liqoctl install kind --cluster-name fluidos-provider \
  --set controllerManager.config.resourcePluginAddress=node-rear-controller-grpc.fluidos:2710 \
  --set controllerManager.config.enableResourceEnforcement=true

if [ $edge_ena -eq 1 ]; then
  rm -rf manifests
  echo "Edge worker nodes can use the following information to join the cluster"
  echo "MASTER_NODE_IP: $ADV_ADDR"
  echo "KE_TOKEN: $KE_TOKEN"
fi

