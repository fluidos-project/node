# Installation

A quick script for installing the FLUIDOS Node is available. Currently, the script supports installation on KIND Clusters, with plans to extend support to generic Kubernetes clusters in the near future.

---

**⚠️ ATTENTION:** The script is currently in an experimental phase, so it may not work as expected. If any issues arise, it may be tricky to understand and terminate the script, as many sub-tasks are executed in the background. We are aware of these issues and are actively working to resolve them.

If you want to use a **working and tested script** to test the FLUIDOS Node within a KinD environment, please refer to the [**Testbed**](../../testbed/kind/README.md) section.

---

To execute the script, use the following command:

```bash
cd tools/scripts
. ./setup.sh
```

No options are available through the CLI, but you can choose the installation mode by choosing the right option during the script execution.
The option supported are:

- `1` to install the FLUIDOS Node as the demo testbed through KIND
- `2` to install the FLUIDOS Node in n consumer clusters and m provider clusters through KIND
- `3` to install the FLUIDOS Node in n clusters through their KUBECONFIG files
  **DISCLAIMER:** in this case all your Kubernetes clusters inserted in the script must have at least one node tagged with `node-role.fluidos.eu/worker: "true"` and at least in the provider clusters, you can choose the nodes that exposes their Kubernetes resources with the label `node-role.fluidos.eu/resources: "true"`.

For each option, you can choose to install from either the official remote FLUIDOS repository or the local repository, by building it at the moment.
