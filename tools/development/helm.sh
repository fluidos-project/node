#!/bin/bash

# Check that there are at least 3 arguments (username, version, at least one component)
if [[ "$#" -lt 3 ]]; then
    echo "Syntaxt error: insufficient arguments."
    echo "Use: $0 DOCKER_USERNAME VERSIONE component [component2 ...]"
    exit 1
fi

DOCKER_USERNAME="$1"
VERSION="$2"

# Remove the first two arguments (username and version) to handle only components and images
shift 2

# Associative array to associate components to their corresponding Helm values
declare -A COMPONENT_MAP
COMPONENT_MAP["rear-controller"]="rearController.imageName"
COMPONENT_MAP["rear-manager"]="rearManager.imageName"
COMPONENT_MAP["local-resource-manager"]="localResourceManager.imageName"

# Initialize a variable to store the --set options
IMAGE_SET_STRING=""

# Iterates over the arguments passed to the script
for component in "$@"; do
    # Check that the component is valid
    if [[ -z "${COMPONENT_MAP[$component]}" ]]; then
        echo "Error: component '$component' not recognized."
        continue
    fi
    
    helm_key="${COMPONENT_MAP[$component]}"
    # Build the --set string using the map
    IMAGE_SET_STRING="$IMAGE_SET_STRING --set $helm_key=$DOCKER_USERNAME/$component"
done

export KUBECONFIG=../../testbed/kind/consumer/config

# Delete all the resources
kubectl delete solvers.nodecore.fluidos.eu -n fluidos --all
kubectl delete discoveries.advertisement.fluidos.eu -n fluidos --all
kubectl delete reservations.reservation.fluidos.eu -n fluidos --all
kubectl delete contracts.reservation.fluidos.eu -n fluidos --all
kubectl delete peeringcandidates.advertisement.fluidos.eu -n fluidos --all
kubectl delete transactions.reservation.fluidos.eu -n fluidos --all
kubectl delete allocations.nodecore.fluidos.eu -n fluidos --all
kubectl delete flavours.nodecore.fluidos.eu -n fluidos --all

kubectl apply -f ../../deployments/node/crds

# Run the helm command
helm get values node -n fluidos > consumer_values.yaml
helm uninstall node -n fluidos
eval "helm install node ../../deployments/node -n fluidos -f consumer_values.yaml $IMAGE_SET_STRING" --set tag="$VERSION"
rm consumer_values.yaml

export KUBECONFIG=../../testbed/kind/provider/config

# Delete all the resources
kubectl delete solvers.nodecore.fluidos.eu -n fluidos --all
kubectl delete discoveries.advertisement.fluidos.eu -n fluidos --all
kubectl delete reservations.reservation.fluidos.eu -n fluidos --all
kubectl delete contracts.reservation.fluidos.eu -n fluidos --all
kubectl delete peeringcandidates.advertisement.fluidos.eu -n fluidos --all
kubectl delete transactions.reservation.fluidos.eu -n fluidos --all
kubectl delete allocations.nodecore.fluidos.eu -n fluidos --all
kubectl delete flavours.nodecore.fluidos.eu -n fluidos --all

kubectl apply -f ../../deployments/node/crds

# Run the helm command
helm get values node -n fluidos > provider_values.yaml
helm uninstall node -n fluidos
eval "helm install node ../../deployments/node -n fluidos -f provider_values.yaml $IMAGE_SET_STRING" --set tag="$VERSION"
rm provider_values.yaml


