#!/bin/bash

# Function to delete a kind cluster and its config file
delete_kind_cluster_and_config() {
    local cluster_name="$1"
    local config_file="$2"

    # Delete the kind cluster if it exists
    if kind get clusters | grep -q "$cluster_name"; then
        echo "Deleting kind cluster: $cluster_name"
        kind delete cluster --name "$cluster_name"
    else
        echo "Kind cluster $cluster_name does not exist"
    fi

    # Delete the kubeconfig file if it exists
    if [ -f "$config_file" ]; then
        echo "Deleting kubeconfig file: $config_file"
        rm -f "$config_file"
    else
        echo "Kubeconfig file $config_file does not exist"
    fi
}

# Loop through the clusters and delete them
for i in {1..10}; do
    delete_kind_cluster_and_config "fluidos-consumer-$i" "fluidos-consumer-$i-config"
    delete_kind_cluster_and_config "fluidos-provider-$i" "fluidos-provider-$i-config"
done

# Delete the JSON files if they exist
json_files=("fluidos-consumers-clusters.json" "fluidos-providers-clusters.json")

for json_file in "${json_files[@]}"; do
    if [ -f "$json_file" ]; then
        echo "Deleting JSON file: $json_file"
        rm -f "$json_file"
    else
        echo "JSON file $json_file does not exist"
    fi
done

echo "Cleanup completed."
