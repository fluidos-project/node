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


helm install node fluidos/node -n fluidos \
--create-namespace -f ../../quickstart/utils/consumer-values.yaml \
--set networkManager.configMaps.nodeIdentity.ip="LOCAL_K8S_CLUSTER_CP_IP:LOCAL_REAR_PORT"\
--set networkManager.configMaps.providers.local="REMOTE_K8S_CLUSTER_CP_IP:REMOTE_REAR_PORT"\
--wait 
```

Due to the absence of the Network Manager component that enable the auto-discovery among FLUIDOS Nodes, we need to setup some parameters manually to allow Nodes discovery.

Here, the meaning of the various parameters:

- LOCAL_K8S_CLUSTER_CP_IP: The IP address of your local Kubernetes cluster Control Plane
- LOCAL_REAR_PORT: The port on which your local cluster uses the REAR protocol
- REMOTE_K8S_CLUSTER_CP_IP: It's the IP address of the remote Kubernetes cluster to which you want to negotiate Flavors thorugh REAR with.
- REMOTE_REAR_PORT: It's the port on which the remote cluster uses the REAR protocol.
