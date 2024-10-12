#!/usr/bin/bash

# Enable job control
set -m


SCRIPT_PATH="$(realpath "${BASH_SOURCE[0]}")"
SCRIPT_DIR="$(dirname "$SCRIPT_PATH")"

# shellcheck disable=SC1091
source "$SCRIPT_DIR"/utils.sh

# PIDs of the processes in background
pids=()

# Function to handle errors
handle_error() {
    echo "An error occurred. Exiting..."
    for pid in "${pids[@]}"; do
        # Kill all the processes in background
        kill "$pid" 2>/dev/null
    done
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

# Create KIND clusters
# Parameters:
# $1: consumer JSON tmp file
# $2: provider JSON tmp file
create_kind_clusters() {

    # Get consumer JSON tmp file from parameter
    consumer_json=$1

    # Get provider JSON tmp file from parameter
    provider_json=$2

    print_title "Create KIND clusters..."

    # Map of clusters:
    # key: cluster name
    # Value: Dictionary with IP of control plane and kubeconfig file
    unset clusters
    declare -A clusters  
    
    # Get parameter to know  wich clusters configuration follow
    # customkind: n consumers and m providers
    if [ "$3" == "customkind" ]; then
        # Create n consumer cluster and m provider clusters
        # Iterate over consumer cluster creation
        for i in $(seq 1 "$4"); do
            (
                # Cluster name
                name="fluidos-consumer-$i"
                # Print cluster creation information
                echo "Creating cluster $name..."
                # Set the role of the cluster
                role="consumer"
                # Create the cluster
                kind create cluster --name "$name" --config "$SCRIPT_DIR"/../../quickstart/kind/configs/standard.yaml --kubeconfig "$SCRIPT_DIR"/"$name"-config -q
                # Install macvlan plugin to enable multicast node discovery, if required
                if [ "$6" == "true" ]; then
                    num_workers=$(kind get nodes --name fluidos-consumer-1 | grep worker -c)
                    for j in $(seq 1 "$num_workers"); do
                        (
                            docker exec --workdir /tmp "$name"-worker"$([ "$j" = 1 ] && echo "" || echo "$j")" mkdir -p cni-plugins
                            docker exec --workdir /tmp/cni-plugins "$name"-worker"$([ "$j" = 1 ] && echo "" || echo "$j")" curl -LO https://github.com/containernetworking/plugins/releases/download/v1.5.1/cni-plugins-linux-amd64-v1.5.1.tgz
                            docker exec --workdir /tmp/cni-plugins "$name"-worker"$([ "$j" = 1 ] && echo "" || echo "$j")" tar xvfz cni-plugins-linux-amd64-v1.5.1.tgz
                            docker exec --workdir /tmp/cni-plugins "$name"-worker"$([ "$j" = 1 ] && echo "" || echo "$j")" cp macvlan /opt/cni/bin
                            docker exec --workdir /tmp "$name"-worker"$([ "$j" = 1 ] && echo "" || echo "$j")" rm -r cni-plugins
                        )
                    done
                fi
                # Get the IP of the control plane of the cluster
                controlplane_ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$name"-control-plane)                
                # Write the cluster info to a file
                echo "$name: {\"ip\":\"$controlplane_ip\", \"kubeconfig\":\"$SCRIPT_DIR/$name-config\", \"role\":\"$role\"}" >> "$consumer_json"
            ) &
            # Save the PID of the process
            pids+=($!)
        done
        # Iterate over provider clusters creation
        for i in $(seq 1 "$5"); do
            (
                # Cluster name
                name="fluidos-provider-$i"
                # Print cluster creation information
                echo "Creating cluster $name..."
                # Set the role of the cluster
                role="provider"
                # Create the cluster
                kind create cluster --name "$name" --config "$SCRIPT_DIR"/../../quickstart/kind/configs/standard.yaml --kubeconfig "$SCRIPT_DIR"/"$name"-config -q
                # Install macvlan plugin to enable multicast node discovery, if required
                if [ "$6" == "true" ]; then
                    num_workers=$(kind get nodes --name fluidos-provider-1 | grep worker -c)
                    for j in $(seq 1 "$num_workers"); do
                        (
                            docker exec --workdir /tmp "$name"-worker"$([ "$j" = 1 ] && echo "" || echo "$j")" mkdir -p cni-plugins
                            docker exec --workdir /tmp/cni-plugins "$name"-worker"$([ "$j" = 1 ] && echo "" || echo "$j")" curl -LO https://github.com/containernetworking/plugins/releases/download/v1.5.1/cni-plugins-linux-amd64-v1.5.1.tgz
                            docker exec --workdir /tmp/cni-plugins "$name"-worker"$([ "$j" = 1 ] && echo "" || echo "$j")" tar xvfz cni-plugins-linux-amd64-v1.5.1.tgz
                            docker exec --workdir /tmp/cni-plugins "$name"-worker"$([ "$j" = 1 ] && echo "" || echo "$j")" cp macvlan /opt/cni/bin
                            docker exec --workdir /tmp "$name"-worker"$([ "$j" = 1 ] && echo "" || echo "$j")" rm -r cni-plugins
                        )
                    done
                fi
                # Get the IP of the control plane of the cluster
                controlplane_ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$name"-control-plane)         
                 # Write the cluster info to a file
                echo "$name: {\"ip\":\"$controlplane_ip\", \"kubeconfig\":\"$SCRIPT_DIR/$name-config\", \"role\":\"$role\"}" >> "$provider_json"
            ) &
            # Save the PID of the process
            pids+=($!)
        done

        # Wait for all the processes to finish
        for pid in "${pids[@]}"; do
            wait "$pid"
        done

    else
        echo "Invalid parameter."
        exit 1
    fi

    print_title "KIND clusters created successfully."

    # Return the clusters
    echo "${clusters[@]}"
}

# Get clusters from KUBECONFIG files
# Parameters:
# $1: consumer JSON tmp file
# $2: provider JSON tmp file
get_clusters() {

    # Get consumer JSON tmp file from parameter
    consumer_json=$1

    # Get provider JSON tmp file from parameter
    provider_json=$2

    print_title "CONSUMER CLUSTERS"

    insert_clusters "$consumer_json" "consumer"

    print_title "PROVIDER CLUSTERS"

    insert_clusters "$provider_json" "provider"
}

# Insert clusters into proper file
# Parameters:
# $1: JSON file
# $2: role
insert_clusters() {

    # Get JSON file from parameter
    json_file=$1

    # Get role from parameter
    role=$2

    while true; do
        # Ask the user for the KUBECONFIG file path
        echo "Insert the KUBECONFIG file path of a $role cluster or press Enter to exit:"
        read -r kubeconfig_path

        # Exit the cycle if the user press Enter
        if [ -z "$kubeconfig_path" ]; then
            break
        fi

        # Check if the file exists
        if [ ! -f "$kubeconfig_path" ]; then
            echo "Errore: il file $kubeconfig_path non esiste."
            continue
        fi

        # Validate KUBECONFIG
        if ! kubectl config view --kubeconfig="$kubeconfig_path" &> /dev/null; then
            echo "Error: file $kubeconfig_path is not a valid KUBECONFIG file."
            continue
        fi

        # Get name of the cluster
        cluster_name=$(kubectl config view --kubeconfig="$kubeconfig_path" -o=jsonpath='{.current-context}')

        # Get the IP of a random node in the cluster
        node_ip=$(kubectl get nodes --kubeconfig="$kubeconfig_path" -o jsonpath='{.items[*].status.addresses[?(@.type=="InternalIP")].address}' | awk '{print $1}')

        # Save the cluster info into json_file file
        echo "$cluster_name: {\"ip\":\"$node_ip\", \"kubeconfig\":\"$kubeconfig_path\", \"role\":\"$role\"}" >> "$json_file"
    done

}