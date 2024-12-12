#!/usr/bin/bash

# Check if provider parameter is provided
if [ -z "$1" ]; then
  echo "No provider specified. Please provide a cloud provider (aws, azure, gcp, etc.)."
  exit 1
fi


# Get the provider parameter
# Get the provider parameter
PROVIDER=$1

# Get the cluster name
CLUSTER_NAME=$2

# Get the Kubeconfig
KUBECONFIG_LIQO=$3

LIQOCTL_PATH=$4

# Print Liqo version
$LIQOCTL_PATH version --client

# Install Liqo based on the provider
$LIQOCTL_PATH install "$PROVIDER" --cluster-id "$CLUSTER_NAME" --kubeconfig "$KUBECONFIG_LIQO" || { echo "Failed to install Liqo for provider: $PROVIDER"; exit 1; }
# liqoctl install "$PROVIDER" || { echo "Failed to install Liqo for provider: $PROVIDER"; exit 1; }

echo "Liqo installation for provider $PROVIDER completed successfully."
