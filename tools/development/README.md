# FLUIDOS Node: Development Toolkit & Guide 🛠️

## Introduction

Welcome to the FLUIDOS Node Development Toolkit and Guide. This comprehensive guide provides a detailed walkthrough on building, testing, and contributing to the FLUIDOS Node.

⚠️ **Note:** This toolkit facilitates the development of your custom FLUIDOS Node version. There are several alternatives that might provide a more efficient development workflow, such as running FLUIDOS Node components directly on your host machine.

## Prerequisites

Ensure the following tools are installed and running:

1. [Docker](https://docs.docker.com/get-docker/)
2. [KIND (Kubernetes in Docker)](https://kind.sigs.k8s.io/docs/user/quick-start/)
3. [Helm](https://helm.sh/docs/intro/install/)
4. [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

## Setup

Follow these steps for initial setup:

1. Clone this repository.
2. Set up the [Testbed for KIND](../../testbed/kind/README.md).

⚠️ **Note:** This Toolkit is **stricly** related to the [Testbed for KIND](../../testbed/kind/README.md). If you are using a different environment, you might need to adapt the scripts for your needs.

## Development

### Environment Preparation

We utilize a `Makefile` to automate common tasks. Refer to the [Makefile](../../Makefile) for a comprehensive list of commands.

Key commands include:

- **API Updates:** After modifying the `/api` directory, regenerate the manifests using:
  
```bash
make manifests
```

- **RBAC Updates:** If code changes necessitate new RBAC rules, apply the correct kubebuilder annotation in the code. Subsequently, update the RBAC rules using:

```bash
make rbac
```

Additional details on kubebuilder annotations can be found [here](https://book.kubebuilder.io/reference/markers/crd.html).

- **General Updates (Recommended):** After making code modifications, use the following command to refresh the manifests, update RBAC rules, and format the code:

```bash
make generate
```

### Build and Run your own FLUIDOS Node version

🔶 For the **first build** of your FLUIDOS Node version, only proceed with steps 1 and 2.

For subsequent builds, proceed with steps 1 and 3, bypassing step 2.

Initially, navigate to the `tools/development` directory:

```bash
cd tools/development
```

#### 1. Image Building & Loading into KIND 📦

Build your code and load images into KIND:

```bash
. ./build.sh <docker_namespace> <version> [<component> ...]
```

Parameters:

- `<docker_namespace>`: Docker username (e.g., `<docker_namespace>/<component>`).
- `<version>`: Component version (e.g., `0.0.1`).
- `<component>`: Specific component to build (e.g., `local-resource-manager`). To build all components, omit this parameter.

Components:

- `local-resource-manager`
- `rear-manager`
- `rear-controller`

#### 2.  Installing Your Custom FLUIDOS Node Helm Chart 🛠️

Upgrade the Helm chart with your custom FLUIDOS Node images:

```bash
. .helm.sh <docker_namespace> <version> <component> [<component> ...]
```

Parameters mirror those in Step 1. Note that the list of components in this step should match those from the image build in the previous step. If no component was specified previously, list all components here.

#### 3. Resource & CR Cleanup 🧹

Purge FLUIDOS resources and CRs:

```bash
. ./purge.sh
```
