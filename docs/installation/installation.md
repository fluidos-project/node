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

If you want to use a **working and tested script** to test the FLUIDOS Node within a KinD environment, please refer to the [**Testbed**](../../testbed/kind/README.md) section. -->

<!-- --- -->

## Testbed installation

To execute the script, use the following command:

```bash
cd tools/scripts
. ./setup.sh
```

No options are available through the CLI, but you can choose the installation mode by choosing the right option during the script execution.
The option supported are:

- `1` to install the FLUIDOS Node as the demo testbed through KIND

- `2` to install the FLUIDOS Node in n consumer clusters and m provider clusters through KIND

For both options, you can choose to install from either the official remote FLUIDOS repository or the local repository, building all the components locally.

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
./install-liqo.sh <provider> <cluster-name> $KUBECONFIG
```

Please, note that you need to pass a few parameters.

- "provider": this parameter depends on your Kubernetes installation.
    We currently test it on the following providers:
    1. kubeadm
    2. k3s
    3. kind
- "cluster-name": this is the name you want to give to your Liqo local cluster (e.g.: `fluidos-turin-1`)

- $KUBECONFIG: it is the typical environment variable that points to the path of your Kubernetes cluster configuration.

For more information, check out [Liqo official documentation](https://docs.liqo.io/en/v0.10.3/installation/install.html#install-with-liqoctl) for all supported providers.

**DISCLAIMER:** before going ahead, ensure that at least one node is tagged with `node-role.fluidos.eu/worker: "true"` and, if acting as a provider, choose the nodes that exposes their Kubernetes resources with the label `node-role.fluidos.eu/resources: "true"`.

Once we have Liqo running, we can install the FLUIDOS Node component via helm:

```bash
helm repo add fluidos https://fluidos-project.github.io/node/


helm upgrade --install node fluidos/node \
    -n fluidos --version "$FLUIDOS_VERSION" \
    --create-namespace -f consumer-values.yaml \
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
