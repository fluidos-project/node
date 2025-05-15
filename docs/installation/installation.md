# Installation

This section shows the installation process of the FLUIDOS Node components.

There are two ways to install FLUIDOS Node components:

1. Launching the testbed environment with a script.

2. Installing the FLUIDOS Node manually on your Kubernetes cluster.

If it's your first time with FLUIDOS Node components, we suggest the testbed environment via script. This allows you to install as many kind clusters as you want to run your experiments.

However, if you want to test FLUIDOS Node on your cluster already setup, we suggest you the manual installation.

<!-- A quick script for installing the FLUIDOS Node is available. Currently, the script supports installation on KIND Clusters, with plans to extend support to generic Kubernetes clusters in the near future. -->

<!-- --- -->
<!-- 
**⚠️ ATTENTION:** The script is currently in an experimental phase, so it may not work as expected. If any issues arise, it may be tricky to understand and terminate the script, as many sub-tasks are executed in the background. We are aware of these issues and are actively working to resolve them.
-->

<!-- --- -->

## Prerequisites

Below are the required tools, along with the versions used by the script:

- [Docker](https://docs.docker.com/get-docker/) v28.1.1
- [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) v1.33.0
- [KIND](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) v0.27.0
- [Helm](https://helm.sh/docs/intro/install/) v3.17.3
- [Liqo CLI tool](https://docs.liqo.io/en/stable/installation/liqoctl.html) v1.0.0

> **Note** The installation script will automatically check if these tools are installed and will ask for your confirmation to install them if they are missing. It will install each tool using a fixed version, except for Docker, which will be installed at the stable version. After Docker installation, an additional CLI command will be required to ensure its proper functionality.

## Common issues with KIND

This setup leverages KIND (Kubernetes IN Docker) to quickly provision a reliable testbed for evaluating the FLUIDOS Node architecture.

When running multiple KIND clusters, certain issues may arise—particularly related to swap memory usage and system-level resource limits. To mitigate these problems, it is strongly recommended to execute the following commands prior to starting the installation:

1. `sudo swapoff -a`
2. `sudo sysctl fs.inotify.max_user_instances=8192`
3. `sudo sysctl fs.inotify.max_user_watches=524288`

## Testbed installation

### What will be installed

The script will create two different types of Kubernetes clusters, each consisting of 3 nodes, as defined in the files located in the `quickstart/kind` directory:

- **fluidos-consumer**: This cluster (also known as a FLUIDOS node) will act as a consumer of FLUIDOS resources. It will use the REAR protocol to communicate with the provider cluster, retrieve available Flavors, and reserve the one that best matches the solver’s request, proceeding to purchase it.

- **fluidos-provider**: This cluster (also known as a FLUIDOS node) will act as a provider of FLUIDOS resources. It will offer its available Flavors in response to requests from the consumer, managing their reservation and sale accordingly.

### Installation

1. Clone the repository

    ```sh
    git clone https://github.com/fluidos-project/node.git
    ```

2. Move into the KIND Example folder

    ```sh
    cd node/tools/scripts
    ```

3. Launch the `setup.sh` script

    ```sh
    ./setup.sh
    ```

4. No command-line arguments are currently supported; instead, the installation mode is selected interactively at the beginning of the script execution. The available options are:

    - `1` Install the FLUIDOS Node using the demo testbed (one consumer and one provider cluster) via KIND.

    - `2` Install the FLUIDOS Node using a custom setup with n consumer clusters and m provider clusters via KIND.

    For both options, you will be prompted to choose:

    - Whether to install from the official remote FLUIDOS repository or use local repositories and build all components locally.

    - Whether to enable resource auto-discovery.

    - Whether to enable LAN node discovery.

    An example prompt flow is shown below:

    ```sh
    1. Use demo KIND environment (one consumer and one provider)
    2. Use a custom KIND environment with n consumer and m provides
    Please enter the number of the option you want to use:
    1
    Do you want to use local repositories? [y/n] y
    Do you want to enable resource auto discovery? [y/n] y
    Do you want to enable LAN node discovery? [y/n] y
    ```

5. At the beginning of the script execution, a check is performed to ensure all [required tools](#prerequisites) are installed. If any dependencies are missing, the script will prompt you for confirmation before proceeding with their automatic installation.
    > **Note** The tools will be installed assuming Linux as the operating system. If you are not using Linux, you will need to install them manually.

    If Docker is not installed and you choose to install it, the script will terminate after its installation. This is necessary because the Docker group must be reloaded in order to use Docker commands without sudo. You can achieve this by running:

    ```sh
    newgrp docker
    ```

    This command opens a new shell with updated group permissions. After executing it, simply restart the installation script.

6. After executing the script, you can verify the status of the pods in the consumer cluster using the following commands:

    ```sh
    export KUBECONFIG=fluidos-consumer-1-config
    kubectl get pods -n fluidos
    ```

    To inspect the provider cluster, use its corresponding kubeconfig file:

    ```sh
    export KUBECONFIG=fluidos-provider-1-config
    kubectl get pods -n fluidos
    ```

    Alternatively, to avoid switching the KUBECONFIG environment variable each time, you can directly specify the configuration file path and context when using kubectl. The paths to the generated configuration files are displayed at the end of the script execution:

    ```sh
    kubectl get pods --kubeconfig "$PWD/fluidos-consumer-1-config" --context kind-fluidos-consumer -n fluidos
    ```

    This approach enables seamless monitoring of both consumer and provider clusters without needing to re-export environment variables manually.

7. You should see 4 pods running on the `fluidos-consumer` cluster and 4 pods running on the `fluidos-provider` cluster:

    - `node-local-resource-manager-<random>`
    - `node-network-manager-<random>`
    - `node-rear-controller-<random>`
    - `node-rear-manager-<random>`

### Usage

In this section, we will guide you through interacting with the FLUIDOS Node at a high level. If you prefer to interact with the FLUIDOS Node using its Custom Resource Definitions (CRDs), please refer to the [Low-Level Usage](../../docs/usage/usage.md) section.

Let’s start by deploying an example `solver` Custom Resource (CR) on the `fluidos-consume`r cluster.

1. Open a new terminal on the repo and move into the `deployments/node/samples` folder

    ```sh
    cd deployments/node/samples
    ```

2. Set the `KUBECONFIG` environment variable to the `fluidos-consumer` cluster

    ```sh
    export KUBECONFIG=../../../tools/scripts/fluidos-consumer-1-config
    ```

3. Deploy the `solver` CR

    ```sh
    kubectl apply -f solver.yaml
    ```

    > **Note**
    > Please review the **architecture** field and change it to **amd64** or **arm64** according to your local machine architecture.

4. Check the result of the deployment

    ```sh
    kubectl get solver -n fluidos
    ```

    The result should be something like this:

    ```sh
    NAME            INTENT ID       FIND CANDIDATE   RESERVE AND BUY   PEERING   STATUS   MESSAGE                               AGE
    solver-sample   intent-sample   true             true              true      Solved   Solver has completed all the phases   83s
    ```

5. Other resources have been created and can be inspected using the following commands:

    ```sh
    kubectl get flavors.nodecore.fluidos.eu -n fluidos
    kubectl get discoveries.advertisement.fluidos.eu -n fluidos
    kubectl get reservations.reservation.fluidos.eu -n fluidos
    kubectl get contracts.reservation.fluidos.eu -n fluidos
    kubectl get peeringcandidates.advertisement.fluidos.eu -n fluidos
    kubectl get transactions.reservation.fluidos.eu -n fluidos
    ```

6. The infrastructure for resource sharing has been established. A demo namespace should now be created in the fluidos-consumer cluster:

    ```sh
    kubectl create namespace demo
    ```

    The namespace can then be offloaded to the fluidos-provider cluster using the following command:

    ```sh
    liqoctl offload namespace demo --pod-offloading-strategy Remote
    ```

    Once the namespace has been offloaded, any workload can be deployed within it. As an example, the provided Kubernetes deployment can be used:

    ```sh
    kubectl apply -f nginx-deployment.yaml -n demo
    ```

Another example involves deploying the `solver-service`, which requests a provider with a database service.

1. Deploy the mock database service in the provider cluster.

    ```sh
    export KUBECONFIG=../../../tools/scripts/fluidos-consumer-1-config
    kubectl apply -f service-blueprint-db.yaml
    ```

2. Deploy the `solver` CR

    ```sh
    kubectl apply -f solver-service.yaml
    ```

3. Check the result of the deployment

    ```sh
    kubectl get solver -n fluidos
    ```

    The result should be something like this:

    ```sh
    NAME                    INTENT ID               FIND CANDIDATE   RESERVE AND BUY   PEERING   STATUS   MESSAGE                               AGE
    solver-sample-service   intent-sample-service   true             true              true      Solved   Solver has completed all the phases   83s
    ```

### Clean Development Environment

If you installed FLUIDOS Node on kind clusters and want to clean up the entire testbed setup, run the following commands:

```bash
cd ../../tools/scripts
./clean-dev-env.sh
```

This script will delete both the kind clusters and their corresponding kubeconfig files.

## Manual installation

Please, make sure you have [helm](https://helm.sh/docs/intro/install/) installed.

To install the FLUIDOS Node on your Kubernetes cluster already up and running, you must ensure you have [Liqo](https://liqo.io/) up and running.

If you are not sure you already have Liqo on your Kubernetes cluster, we suggest to check the next sub-section.

### Install Liqo on your Kubernetes cluster

To ensure you have Liqo, please run the following script:

```bash
cd ../../tools/scripts
chmod +x install_liqo.sh
./install_liqo.sh <provider> <cluster-name> $KUBECONFIG <liqoctl PATH>
```

Please, note that you need to pass a few parameters.

- "provider": this parameter depends on your Kubernetes installation.
    We currently test it on the following providers:
    1. kubeadm
    2. k3s
    3. kind

- "cluster-name": this is the name you want to give to your Liqo local cluster (e.g.: `fluidos-turin-1`)

- $KUBECONFIG: it is the typical environment variable that points to the path of your Kubernetes cluster configuration.

- "liqoctl PATH": it is the path to the liqoctl command. If installed via Liqo guide, liqoctl is sufficient.

For more information, check out [Liqo official documentation](https://docs.liqo.io/en/v1.0.0/installation/install.html#install-with-liqoctl) for all supported providers.

**DISCLAIMER:** before going ahead, ensure that at least one node is tagged with `node-role.fluidos.eu/worker: "true"` and, if acting as a provider, choose the nodes that exposes their Kubernetes resources with the label `node-role.fluidos.eu/resources: "true"`.

Once we have Liqo running, we can install the FLUIDOS Node component via helm:

```bash
helm repo add fluidos https://fluidos-project.github.io/node/


helm upgrade --install node fluidos/node \
    -n fluidos --version "$FLUIDOS_VERSION" \
    --create-namespace -f $PWD/node/quickstart/utils/consumer-values.yaml \
    --set networkManager.configMaps.nodeIdentity.ip="$NODE_IP" \
    --set rearController.service.gateway.nodePort.port="$REAR_PORT" \
    --set networkManager.config.enableLocalDiscovery="$ENABLE_LOCAL_DISCOVERY" \
    --set networkManager.config.address.thirdOctet="$THIRD_OCTET" \
    --set networkManager.config.netInterface="$NET_INTERFACE" \
    --wait \
    --debug \
    --v=2
```

Here, the meaning of the various parameters:

- FLUIDOS_VERSION: The FLUIDOS Node version to be installed
- NODE_IP: The IP address of your local Kubernetes cluster Control Plane
- REAR_PORT: The port on which your local cluster uses the REAR protocol
- ENABLE_LOCAL_DISCOVERY: A flag that enables the Network Manager, which is the component advertising the local FLUIDOS Node into a LAN
- THIRD_OCTET: This is the third byte of the IP address used by Multus CNI for sending broadcast messages into the LAN. **Warning**: this parameters should be different for each FLUIDOS Node to be working (e.g. 1 for the 1st cluster, 2 for the 2nd cluster, etc.)
- NET_INTERFACE: The host network interface that Multus binds to

### Broker CR creation

To enable the Network Manager to discover FLUIDOS Nodes outside a LAN, you need to configure and apply a Broker CR.
The `broker-creation.sh` script simplifies this process by guiding you through the creation of the Broker YAML file.

What you need:

- Kubeconfig PATH

- Broker's name             (custom name of your choice)
- Broker's server address
- Client certificate        (.pem)
- Client private key        (.pem)
- Broker's Root certificate (.pem)
- Role                      (publisher ^ subscriber ^ both)

Two Kubernetes Secrets are created from the certificates and key, one containing the client cert and private key, the other containing the root certificate of the remote broker.
The script will create and apply the .yaml.
To inspect the new Broker: `kubectl describe broker my-broker -n fluidos`
Once applied, the Network Manager Reconcile process starts, enabling message exchange.
