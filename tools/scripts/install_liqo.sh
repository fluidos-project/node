#!/bin/bash


# Function to check if liqoctl is installed
check_and_install_liqoctl() {
  if ! command -v liqoctl &> /dev/null; then
    echo "liqoctl not found. Installing liqoctl..."
    # Example installation command for liqoctl, you may need to update this based on the official installation instructions
    curl -sL https://get.liqo.io | bash || { echo "Failed to install liqoctl"; exit 1; }
    echo "liqoctl installed successfully."
  else
    echo "liqoctl is already installed."
  fi
}

# Check if provider parameter is provided
if [ -z "$1" ]; then
  echo "No provider specified. Please provide a cloud provider (aws, azure, gcp, etc.)."
  exit 1
fi

check_and_install_liqoctl

# Get the provider parameter
PROVIDER=$1

control_plane_node=$(kubectl get nodes -l node-role.kubernetes.io/control-plane -o jsonpath='{.items[0].metadata.name}')
cluster_name=${control_plane_node%-control-plane}


# Install Liqo based on the provider
liqoctl install "$PROVIDER" --cluster-name "$cluster_name" || { echo "Failed to install Liqo for provider: $PROVIDER"; exit 1; }
# liqoctl install "$PROVIDER" || { echo "Failed to install Liqo for provider: $PROVIDER"; exit 1; }

echo "Liqo installation for provider $PROVIDER completed successfully."
