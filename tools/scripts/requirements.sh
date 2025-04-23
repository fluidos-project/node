#!/usr/bin/bash

SCRIPT_PATH="$(realpath "${BASH_SOURCE[0]}")"
SCRIPT_DIR="$(dirname "$SCRIPT_PATH")"

# shellcheck disable=SC1091
source "$SCRIPT_DIR"/utils.sh

# Install KIND function
function install_kind() {
    print_title "Install kind..."
    # Check AMD64 or ARM64
    ARCH=$(uname -m)
    if [ "$ARCH" == "x86_64" ]; then
        ARCH="amd64"
    elif [ "$ARCH" == "aarch64" ]; then
        ARCH="arm64"
    else
        echo "Unsupported architecture."
        exit 1
    fi
    # Install kind if AMD64
    if [ "$ARCH" == "amd64" ]; then
        echo "Install kind AMD64..."
        [ "$(uname -m)" = x86_64 ] && curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.21.0/kind-linux-amd64
        chmod +x kind
        sudo mv kind /usr/local/bin/kind
    elif [ "$ARCH" == "arm64" ]; then
        echo "Install kind ARM64..."
        [ "$(uname -m)" = aarch64 ] && curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.21.0/kind-linux-arm64
        chmod +x kind
        sudo mv kind /usr/local/bin/kind
    fi
    print_title "Kind installed successfully."
}

# Install docker function
function install_docker() {
    print_title "Install docker..."
    # Add Docker's official GPG key:
    sudo apt-get update
    sudo apt-get install ca-certificates curl
    sudo install -m 0755 -d /etc/apt/keyrings
    sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
    sudo chmod a+r /etc/apt/keyrings/docker.asc
    # Add the repository to Apt sources:
    # shellcheck disable=SC1091
    echo \
    "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
    $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
    sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
    sudo apt-get update
    sudo apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
    print_title "Docker installed successfully."
}

# Check docker function
function check_docker() {
    print_title "Check docker..."
    if ! docker -v; then
        echo "Please install docker first."
        return 1
    fi
}

# Install Kubectl function
function install_kubectl() {
    print_title "Install kubectl..."
    # Check AMD64 or ARM64
    ARCH=$(uname -m)
    if [ "$ARCH" == "x86_64" ]; then
        ARCH="amd64"
    elif [ "$ARCH" == "aarch64" ]; then
        ARCH="arm64"
    else
        echo "Unsupported architecture."
        return 1
    fi
    # Install kubectl if AMD64
    if [ "$ARCH" == "amd64" ]; then
        echo "Install kubectl AMD64..."
        curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" 
        sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
    elif [ "$ARCH" == "arm64" ]; then
        echo "Install kubectl ARM64..."
        curl -LO "https://dl.k8s.io/release/v1.21.0/bin/linux/arm64/kubectl"
        sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
    fi
    print_title "Kubectl installed successfully."
}

# Check Kubectl function
function check_kubectl() {
    print_title "Check kubectl..."
    if ! kubectl version --client; then
        # Ask the user if they want to install kubectl
        read -r -p "Do you want to install kubectl? (y/n): " install_kubectl
        if [ "$install_kubectl" == "y" ]; then
            install_kubectl
        else
            echo "Please install kubectl first. Exiting..."
            return 1
        fi
    fi
}

# Install Helm function
function install_helm() {
    print_title "Install helm..."
    curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
    chmod 700 get_helm.sh
    ./get_helm.sh
    print_title "Helm installed successfully."
}

# Check Helm function
function check_helm() {
    print_title "Check helm..."
    helm version
    if ! helm version; then
        # Ask the user if they want to install helm
        read -r -p "Do you want to install helm? (y/n): " install_helm
        if [ "$install_helm" == "y" ]; then
            install_helm
        else
            echo "Please install helm first. Exiting..."
            exit 1
        fi
    fi
}

# Install liqoctl function
function install_liqoctl() {
    print_title "Install liqo..."
    # Check AMD64 or ARM64
    ARCH=$(uname -m)
    if [ "$ARCH" == "x86_64" ]; then
        ARCH="amd64"
    elif [ "$ARCH" == "aarch64" ]; then
        ARCH="arm64"
    else
        echo "Unsupported architecture."
        exit 1
    fi
    # Install liqoctl if AMD64
    if [ "$ARCH" == "amd64" ]; then
        echo "Install liqoctl AMD64..."
        curl --fail -LS "https://github.com/liqotech/liqo/releases/download/v0.10.1/liqoctl-linux-amd64.tar.gz" | tar -xz
        sudo install -o root -g root -m 0755 liqoctl /usr/local/bin/liqoctl
    elif [ "$ARCH" == "arm64" ]; then
        echo "Install liqoctl ARM64..."
        curl --fail -LS "https://github.com/liqotech/liqo/releases/download/v0.10.1/liqoctl-linux-arm64.tar.gz" | tar -xz
        sudo install -o root -g root -m 0755 liqoctl /usr/local/bin/liqoctl
    fi
    print_title "Liqo installed successfully."
}

# Check liqoctl function
function check_liqoctl() {
    print_title "Check liqoctl..."    
    check_and_install_liqoctl
}

# Function to check if liqoctl is installed
check_and_install_liqoctl() {
  if ! command -v liqoctl &> /dev/null; then
    echo "liqoctl not found. Installing liqoctl..."
    # Example installation command for liqoctl, you may need to update this based on the official installation instructions
    install_liqo_not_stable_version
    echo "liqoctl installed successfully."
  else
    # Check the version of the client version of liqo
    CLIENT_VERSION=$(liqoctl version 2>&1 | grep -oP 'Client version: \K\S+')
    if [ -z "$CLIENT_VERSION" ]; then
      echo "Failed to retrieve liqoctl client version"
      exit 1
    else
      echo "liqoctl client version: $CLIENT_VERSION"
      # TODO: Update the version check based on the stable version
      # Version currently used is an unstable version, rc.3
      if [ "$CLIENT_VERSION" != "v1.0.0-rc.3" ]; then
        echo "liqoctl is not installed at the desired version of v1.0.0-rc.3. Installing liqoctl..."
        install_liqo_not_stable_version
      else 
        echo "liqoctl is already installed at the version $CLIENT_VERSION."
      fi
    fi
  fi
}

install_liqo_not_stable_version() {
  # Delete if exists the temporary liqo folder
  rm -rf /tmp/liqo
  # Clone Liqo repository to local tmp folder
  git clone --depth 1 --branch v1.0.0-rc.3 https://github.com/liqotech/liqo.git /tmp/liqo || { echo "Failed to clone Liqo repository"; exit 1; }
  make -C /tmp/liqo ctl || { echo "Failed to install Liqo"; exit 1; }
  echo "Liqo compiled successfully in /tmp/liqo."
  # Create temporary alias for liqoctl to make it available in the current shell
  alias liqoctl=/tmp/liqo/liqoctl
  echo "liqoctl alias created to /tmp/liqo/liqoctl for the current shell."

  shopt -s expand_aliases
}

# Install jq function
function install_jq() {
    print_title "Install jq..."
    sudo apt-get install jq
    print_title "jq installed successfully."
}

# Check jq function
function check_jq() {
    if ! jq --version; then
        # Ask the user if they want to install jq
        read -r -p "Do you want to install jq? (y/n): " install_jq
        if [ "$install_jq" == "y" ]; then
            install_jq
        else
            echo "Please install jq first. Exiting..."
            exit 1
        fi
    fi
}

# Check all the tools
function check_tools() {
    print_title "Check all the tools..."
    check_jq
    check_docker
    check_kubectl
    check_helm
    check_liqoctl
}