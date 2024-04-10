#!/usr/bin/bash

SCRIPT_PATH=$(realpath "${BASH_SOURCE[0]}")
SCRIPT_DIR="$(dirname "$SCRIPT_PATH")"

# shellcheck disable=SC1091
source "$SCRIPT_DIR"/utils.sh

declare -A providers_ips

# PIDs of the processes in background
pids=()

# Function to handle errors
handle_error() {
    echo "An error occurred. Exiting..."
    for pid in "${pids[@]}"; do
        # Kill all the processes in background
        kill "$pid" 2>/dev/null
    done
    read -r -p "All the processes in background have been killed. Press enter to exit."
    return 1
}

# Function to handle exit
handle_exit() {
    echo "Exiting..."
    for pid in "${pids[@]}"; do
        # Kill all the processes in background
        kill "$pid" 2>/dev/null
    done
    # Ask the user if really wants to exit
    read -r -p "Do you really want to exit? [y/N] " answer
    if [ "$answer" == "y" ]; then
        return 0
    fi
}

# Build and load the docker image
build_and_load() {
    local COMPONENT="$1"
    local NAMESPACE="$2"
    local VERSION="$3"
    # Build the docker image
    docker build -q -f "$SCRIPT_DIR"/../../build/common/Dockerfile --build-arg COMPONENT="$COMPONENT" -t "$NAMESPACE"/"$COMPONENT":"$VERSION" "$SCRIPT_DIR"/../../

    echo "Docker image $NAMESPACE/$COMPONENT:$VERSION built"
    # For each cluster, load the docker image
    for cluster in "${!clusters[@]}"; do
        kind load docker-image "$NAMESPACE"/"$COMPONENT":"$VERSION" --name="$cluster"
    done
}

# Install remote components function
# Parameters:
# $1: consumer JSON tmp file
# $2: provider JSON tmp file
# $3: local repositories boolean
# $4: local resource manager boolean
# Return: none
function install_components() {

    unset clusters
    declare -A clusters

    unset providers_ips
    declare -A providers_ips

    # Get consumer JSON tmp file from parameter
    consumers_json=$1

    # Get provider JSON tmp file from parameter
    providers_json=$2

    # Get the remote boolean from parameters
    local_repositories=$3

    # Get the local resource manager installation boolean from parameters
    local_resource_manager=$4

    helm repo add fluidos https://fluidos-project.github.io/node/

    consumer_node_port=30000
    provider_node_port=30001

    # Read the results from the files
    while IFS= read -r line; do
        echo 
        name=$(echo "$line" | cut -d: -f1)
        info=$(echo "$line" | cut -d: -f2-)
        clusters["$name"]=$info
    done < "$consumers_json"

    while IFS= read -r line; do
        name=$(echo "$line" | cut -d: -f1)
        info=$(echo "$line" | cut -d: -f2-)
        clusters["$name"]=$info
    done < "$providers_json"
    
    # Print the clusters
    for cluster in "${!clusters[@]}"; do 
        echo "Cluster: $cluster"
        echo "Value: ${clusters[$cluster]}"
    done

    if [ "$local_repositories" == "true" ]; then
        unset COMPONENT_MAP
        declare -A COMPONENT_MAP
        COMPONENT_MAP["rear-controller"]="rearController.imageName"
        COMPONENT_MAP["rear-manager"]="rearManager.imageName"
        COMPONENT_MAP["local-resource-manager"]="localResourceManager.imageName"
        # Build the image name using the username
        IMAGE_SET_STRING=""
        DOCKER_USERNAME="fluidoscustom"
        VERSION="0.0.1"
        for component in rear-controller rear-manager local-resource-manager; do
            helm_key="${COMPONENT_MAP[$component]}"
            IMAGE_SET_STRING="$IMAGE_SET_STRING --set $helm_key=$DOCKER_USERNAME/$component"
            # Build and load the docker image
            (
                build_and_load $component $DOCKER_USERNAME $VERSION
            ) &
            # Save the PID of the process
            pids+=($!)
        done

        # Wait for each process and if any of them fails, generates a trap to be captured, which kills all the processes and exits
        for pid in "${pids[@]}"; do
            wait "$pid" || handle_error
            echo "Process $pid finished"
        done

        # Reset the pids array
        pids=()
    fi

    # Iterate over the clusters
    for cluster in "${!clusters[@]}"; do

        (
        echo "Cluster is: $cluster"
        echo "Cluster value is: ${clusters[$cluster]}"

        # Create list of providers ip taking all the clusters controlplane IPs from the map and put it ina  string separated by commas
        for provider in "${!clusters[@]}"; do
            # Check if the cluster is not the current one
            # Check if the cluster is a provider
            cluster_role=$(jq -r '.role' <<< "${clusters[$provider]}")
            # Print cluster role
            echo "Cluster role is: $cluster_role"
            if [ "$provider" != "$cluster" ] && [ "$cluster_role" == "provider" ]; then
                # Print the specific cluster informations
                echo "Provider cluster: $provider"
                echo "Value: ${clusters[$provider]}"
                ip_value="${clusters[$provider]}"
                ip=$(jq -r '.ip' <<< "$ip_value")
                # Add the provider port to the IP
                ip="$ip:$provider_node_port"

                if [ -z "${providers_ips[$cluster]}" ]; then
                    providers_ips[$cluster]="$ip"
                else
                    providers_ips[$cluster]="${providers_ips[$cluster]}\,$ip"
                fi
            fi
        done

        echo "Providers IPs for cluster $cluster: ${providers_ips[$cluster]}"

        # Set the KUBECONFIG environment variable taking the value
        export KUBECONFIG
        KUBECONFIG=$(echo "${clusters[$cluster]}" | jq -r '.kubeconfig')

        echo "The KUBECONFIG is $KUBECONFIG"

        # Apply the metrics-server
        kubectl apply -f "$SCRIPT_DIR"/../../quickstart/utils/metrics-server.yaml --kubeconfig "$KUBECONFIG"

        # Wait for the metrics-server to be ready
        echo "Waiting for metrics-server to be ready"
        kubectl wait --for=condition=ready pod -l k8s-app=metrics-server -n kube-system --timeout=300s --kubeconfig "$KUBECONFIG"

        # Decide value file to use based on the role of the cluster
        if [ "$(jq -r '.role' <<< "${clusters[$cluster]}")" == "consumer" ]; then
            # Check if local resouce manager is enabled
            if [ "$local_resource_manager" == "true" ]; then
                value_file="$SCRIPT_DIR/../../quickstart/utils/consumer-values.yaml"
            else
                value_file="$SCRIPT_DIR/../../quickstart/utils/consumer-values-nolrm.yaml"
            fi
            # Get cluster IP and port
            ip_value="${clusters[$cluster]}"
            ip=$(jq -r '.ip' <<< "$ip_value")
            port=$consumer_node_port
        else
            # Check if local resouce manager is enabled
            if [ "$local_resource_manager" == "true" ]; then
                value_file="$SCRIPT_DIR/../../quickstart/utils/provider-values.yaml"
            else
                value_file="$SCRIPT_DIR/../../quickstart/utils/provider-values-nolrm.yaml"
            fi
            # Get cluster IP and port
            ip_value="${clusters[$cluster]}"
            ip=$(jq -r '.ip' <<< "$ip_value")
            port=$provider_node_port
        fi

        # Install the node Helm chart
        # The installation set statically all the other nodes as providers and the current node as the consumer
        echo "Installing node Helm chart in cluster $cluster"
        # If the installation does not use remote repository, the image is used the one built locally
        if [ "$local_repositories" == "true" ]; then
            # If the installation does not use remote repository, the CRDs are applied
            kubectl apply -f "$SCRIPT_DIR"/../../deployments/node/crds --kubeconfig "$KUBECONFIG"
            echo "Installing local repositories in cluster $cluster with local resource manager"
            # Execute command
            # shellcheck disable=SC2086
            helm upgrade --install node $SCRIPT_DIR/../../deployments/node \
            -n fluidos --create-namespace -f $value_file $IMAGE_SET_STRING \
            --set tag=$VERSION \
            --set "networkManager.configMaps.nodeIdentity.ip=$ip:$port" \
            --set "networkManager.configMaps.providers.local=${providers_ips[$cluster]}" \
            --kubeconfig $KUBECONFIG
        else
            echo "Installing remote repositories in cluster $cluster with local resource manager"
            helm upgrade --install node fluidos/node -n fluidos --create-namespace -f "$value_file" \
            --set "networkManager.configMaps.nodeIdentity.ip=$ip:$port" \
            --set 'networkManager.configMaps.providers.local'="${providers_ips[$cluster]}" \
            --kubeconfig "$KUBECONFIG"
        fi

        echo "Installing LIQO in cluster $cluster"
        liqoctl install kind \
        --cluster-name "$cluster" \
        --set controllerManager.config.resourcePluginAddress=node-rear-controller-grpc.fluidos:2710 \
        --set controllerManager.config.enableResourceEnforcement=true \
        --kubeconfig "$KUBECONFIG"
        ) &

        # Save the PID of the process
        pids+=($!)

    done

    # Wait for each process and if any of them fails, generates a trap to be captured, which kills all the processes and exits
    for pid in "${pids[@]}"; do
        wait "$pid" || handle_error
        echo "Process $pid finished"
    done
}