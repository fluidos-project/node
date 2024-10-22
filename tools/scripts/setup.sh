#!/usr/bin/bash

# Enable job control
set -e

# Set traps to handle errors
trap 'handle_error' ERR
# Set trap to handle exit
trap 'handle_exit' INT


SCRIPT_PATH="$(realpath "${BASH_SOURCE[0]}")"
SCRIPT_DIR="$(dirname "$SCRIPT_PATH")"

# shellcheck disable=SC1091
source "$SCRIPT_DIR"/requirements.sh
# shellcheck disable=SC1091
source "$SCRIPT_DIR"/utils.sh
# shellcheck disable=SC1091
source "$SCRIPT_DIR"/environment.sh
# shellcheck disable=SC1091
source "$SCRIPT_DIR"/installation.sh

# Tmp consumer JSON file
consumers_json="$SCRIPT_DIR/fluidos-consumers-clusters.json"

# Tmp provider JSON file
providers_json="$SCRIPT_DIR/fluidos-providers-clusters.json"


# FLUIDOS node installer greetings into the terminal
print_title "Welcome to the FLUIDOS node installer"

echo "We'll now run the installation process for the FLUIDOS node."

# Ask the user what type of enviroment they want to use/create
# Options are:
# 1. Use demo KIND enviroment (one consumer and one provider)
# 2. Use a custom KIND enviroment with n clusters (half consumer, half provider)
# 3. Use personal Kubernetes clusters through KUBECONFIG files (not supported yet)
read -r -p "What type of environment do you want to use? /
1. Use demo KIND environment (one consumer and one provider) /
2. Use a custom KIND environment with n consumer and m provides /
Please enter the number of the option you want to use:
 " environment_type

# Check if the input is a number
if ! [[ $environment_type =~ ^[0-9]+$ ]]; then
    echo "Please enter a number."
    return 1
fi

# Ask the user if they want to use local repositories or the public ones
read -r -p "Do you want to use local repositories? [y/n] " local_repositories

# Check if the input is y or n
if [ "$local_repositories" == "y" ]; then
    # If the enviroment is Kubernetes cluster, the user can't use local repositories
    if [ "$environment_type" -eq 3 ]; then
        # Option not available at the moment
        echo "Option not available at the moment."
        echo "You can't use local repositories with a personal Kubernetes cluster."
        read -r -p "Press any key to continue..."
        return 0
    fi
    local_repositories=true
elif [ "$local_repositories" == "n" ]; then
    local_repositories=false
else
    echo "Invalid option."
    return 1
fi

# Ask the user if they want to use the resource auto discovery from the local resource manager
read -r -p "Do you want to enable resource auto discovery? [y/n] " enable_auto_discovery

# Check if the input is y or n
if [ "$enable_auto_discovery" == "y" ]; then
    enable_auto_discovery=true
elif [ "$enable_auto_discovery" == "n" ]; then
    enable_auto_discovery=false
else
    echo "Invalid option."
    return 1
fi

# Ask the user if they want to use the node discovery from the network resource manager
read -r -p "Do you want to enable LAN node discovery? [y/n] " enable_local_discovery

# Check if the input is y or n
if [ "$enable_local_discovery" == "y" ]; then
    enable_local_discovery=true
elif [ "$enable_local_discovery" == "n" ]; then
    enable_local_discovery=false
else
    echo "Invalid option."
    return 1
fi

# Check requirements with function check_tools from requirements.sh
check_tools

echo "All the tools are installed."

# Check if the input is 1, 2 or 3
if [ "$environment_type" -eq 1 ]; then
    environment_type="customkind"
    installation_type="kind"
    # Call create_kind clusters with parameters and save return value into clusters variable
    create_kind_clusters "$consumers_json" "$providers_json" $environment_type 1 1 $enable_local_discovery
elif [ "$environment_type" -eq 2 ]; then
    environment_type="customkind"
    installation_type="kind"
    # Ask the user how many consumer and provider clusters they want
    read -r -p "How many consumer clusters do you want? " consumer_clusters
    read -r -p "How many provider clusters do you want? " provider_clusters

    # Check if the input is a number
    if ! [[ $consumer_clusters =~ ^[0-9]+$ ]] || ! [[ $provider_clusters =~ ^[0-9]+$ ]]; then
        echo "Please enter a number."
        return 1
    fi

    # Call create_kind clusters with parameters and save return value into clusters variable
    create_kind_clusters "$consumers_json" "$providers_json" $environment_type "$consumer_clusters" "$provider_clusters" $enable_local_discovery
# elif [ "$environment_type" -eq 3 ]; then
#     # Ask the user what Kubernetes clusters they want to use between kubeadm and k3s
#     read -r -p "What type of Kubernetes clusters do you want to use? 
#     1. kubeadm 
#     2. k3s 
#     Please enter the number of the option you want to use:
#     " installation_type
#     if [ "$installation_type" -eq 1 ]; then
#         installation_type="kubeadm"
#     elif [ "$installation_type" -eq 2 ]; then
#         installation_type="k3s"
#     else
#         echo "Invalid option."
#         return 1
#     fi
#     get_clusters "$consumers_json" "$providers_json"
else
    echo "Invalid option."
    return 1
fi

# FLUIDOS node installation
install_components "$consumers_json" "$providers_json" $local_repositories $enable_auto_discovery $installation_type $enable_local_discovery

print_title "Installation completed successfully"

# Print KUBECONFIG files for each cluster
echo "KUBECONFIG files for each cluster:"

# Create cluster variable
unset clusters
declare -A clusters

print_title "Consumer Clusters"
# Read consumers
while IFS= read -r line; do
    echo 
    name=$(echo "$line" | cut -d: -f1)
    info=$(echo "$line" | cut -d: -f2-)
    clusters["$name"]=$info
done < "$consumers_json"

# Print KUBECONFIG files for each consumer cluster
for cluster in "${!clusters[@]}"; do
    KUBECONFIG=$(echo "${clusters[$cluster]}" | jq -r '.kubeconfig')
    echo "$cluster: $KUBECONFIG"
done

unset clusters
declare -A clusters

print_title "Provider Clusters"

# Read consumers
while IFS= read -r line; do
    echo 
    name=$(echo "$line" | cut -d: -f1)
    info=$(echo "$line" | cut -d: -f2-)
    clusters["$name"]=$info
done < "$providers_json"

# Print KUBECONFIG files for each consumer cluster
for cluster in "${!clusters[@]}"; do
    KUBECONFIG=$(echo "${clusters[$cluster]}" | jq -r '.kubeconfig')
    echo "$cluster: $KUBECONFIG"
done

rm "$consumers_json"
rm "$providers_json"