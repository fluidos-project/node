<!-- markdownlint-disable first-line-h1 -->
<p align="center">
<a href="https://www.fluidos.eu/"> <img src="./docs/images/fluidoslogo.png" width="150"/> </a>
<h3 align="center">FLUIDOS Node</h3>
</p>

## What is a FLUIDOS Node?

A FLUIDOS node is a Kubernetes cluster, orchestrated by a single control plane instance, and it can be composed of either a single machine (e.g., an embedded device) or a set of servers (e.g., a datacenter).
Device homogeneity is desired in order to simplify the management (physical servers can be considered all equals, since they feature a similar amount of hardware resources), but it is not requested within a FLUIDOS node. In other words, a FLUIDOS node corresponds to a *Kubernetes cluster*.

A FLUIDOS node handles problems such as orchestrating computing, storage, network resources and software services within the cluster and, thanks to [Liqo](https://liqo.io), can transparently access to resources and services that are running in another (remote) Kubernetes cluster (a.k.a. remote FLUIDOS node).

## What can I find in this repo?

This repository contains the FLUIDOS Node, along with its essential components, such as:

- [**Local ResourceManager**](/docs/implementation/components.md#local-resourcemanager)
- [**Available Resources**](/docs/implementation/components.md#available-resources)
- [**Discovery Manager**](/docs/implementation/components.md#discovery-manager)
- [**Peering Candidates**](/docs/implementation/components.md#peering-candidates)
- [**REAR Manager**](/docs/implementation/components.md#rear-manager)
- [**Contract Manager**](/docs/implementation/components.md#contract-manager)

Please note that this repository is continually updated, with additional components slated for future inclusion.

## Implementation

Want to know more about the implementation? Check out the [**Implementation**](./docs/implementation/implementation.md) section.

## Installation

Want to know how to install a FLUIDOS Node? Check out the [**Installation**](./docs/installation/installation.md) section.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## How to Contribute

Please, refer to the [Contributing](CONTRIBUTING.md) guide on how to contribute.
