<p align="center">
<a href="https://www.fluidos.eu/"> <img src="./docs/images/fluidoslogo.png" width="150"/> </a>
<h3 align="center">WP3 - FLUIDOS Node</h3>
</p>

## What is a FLUIDOS Node?
A FLUIDOS node is orchestrated by a single Kubernetes control plane, and it can be composed of either a single device or a set of devices (e.g., a datacenter). Device homogeneity is desired in order to simplify the management (physical servers can be considered all equals, since they feature a similar amount of hardware resources), but it is not requested within a FLUIDOS node. In other words, a FLUIDOS node corresponds to a *Kubernetes cluster*.

## What can I find in this repo?
This repository contains the FLUIDOS Node, along with its essential components, such as: 

- [**Local ResourceManager**](/docs/implementation/components.md#local-resourcemanager)
- [**Avaialable Resources**](/docs/implementation/components.md#available-resources) 
- [**Discovery Manager**](/docs/implementation/components.md#discovery-manager)
- [**Peering Candidates**](/docs/implementation/components.md#peering-candidates)
- [**REAR Manager**](/docs/implementation/components.md#rear-manager)
- [**Contract Manager**](/docs/implementation/components.md#contract-manager)

Please note that this repository is continually updated, with additional components slated for future inclusion.

## Implementation
Want to know more about the implementation? Check out the [**Implementation Part**](./docs/implementation/implementation.md).

## Installation
Want to know how to install a FLUIDOS Node? Check out the [**Installation Part**](./docs/installation/installation.md).

## License
This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.