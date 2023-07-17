## Components
Let's how each component has been implemented:

- [**Local ResourceManager**](#local-resourcemanager)
- [**Available Resources**](#available-resources)
- [**Discovery Manager**](#discovery-manager)
- [**Peering Candidates**](#peering-candidates)
- [**REAR Manager**](#rear-manager)
- [**Contract Manager**](#contract-manager)

### Local ResourceManager
The **Local Resource Manager** was constructed through the development of a Kubernetes controller. This controller serves the purpose of monitoring the internal resources of individual nodes within a FLUIDOS Node, representing a cluster. Subsequently, it generates a *Flavour Custom Resource (CR)* for each node and stores these CRs within the cluster for further management and utilization.

### Available Resources 
**Available Resources** component is a critical part of the FLUIDOS system responsible for managing and storing Flavours. It consists of two primary data structures:

1. **Free Flavours**: This section is dedicated to storing Flavours, as defined in the REAR component.

2. **Reserved Flavours**: In this section, objects or documents representing various resource allocations are stored, with each represented as a Flavour.

The primary function of Available Resources is to serve as a centralized repository for Flavours. When queried, it provides a list of these resources. Under the hood, Available Resources seamlessly integrates with Kubernetes' `etcd`, ensuring efficient storage and retrieval of data.

This component plays a crucial role in facilitating resource management within the FLUIDOS ecosystem, enabling efficient allocation and utilization of computing resources.
### Discovery Manager
The **Discovery Manager** component within the FLUIDOS system serves as a critical part of the resource discovery process. It operates as a Kubernetes controller, continuously monitoring Discovery Custom Resources (CRs) generated by the Solver Controller.

The primary objectives of the Discovery Manager are as follows:

- **Populating Peering Candidates Table**: The Discovery Manager's primary responsibility is to populate the Peering Candidates table. It achieves this by identifying suitable resources known as "Flavours" based on the initial messages exchanged as part of the REAR protocol.

- **Offering Appropriate Flavours**: If no suitable peering candidates are initially available, the REAR Manager constructs a flavor selector and forwards it to the Discovery Manager. The Discovery Manager, in turn, initiates a LIST_FLAVOURS message, broadcasting it to all known endpoints, including local Nodes, Supernodes, and Catalogs.

The Discovery Manager plays a pivotal role in facilitating resource discovery and allocation within the FLUIDOS ecosystem. It contributes to the seamless orchestration of resources across FLUIDOS Nodes, ensuring efficient utilization and allocation of computing resources.


### Peering Candidates
The **Peering Candidates** component manages a dynamic list of nodes that are potentially suitable for establishing peering connections. This list is continuously updated based on the available resources in the nodes and the requests for flavours from the Discovery Manager.

Under the hood, Peering Candidates seamlessly integrates with Kubernetes' `etcd`, ensuring efficient storage and retrieval of data.

### REAR Manager
The **REAR Manager** plays a pivotal role in orchestrating the service provisioning process. It receives service requests, translates them into resource or service requests, and looks up suitable resources. 

If none are found, it initiates the Discovery, Reservation, and Peering phases. It selects potential peering candidates and can request discovery. The suitable local flavors are identified and shared. If a suitable candidate is found, it triggers the Reservation phase. Resources are allocated, contracts are created, and peering is established.

### Contract Manager
The Contract Manager is in charge of managing the reserve and acquisition of resources. It handles the negotiation and management of resource contracts between nodes. 

When a suitable peering candidate is identified, the Contract Manager initiates the Reservation phase by sending a **RESERVE\_FLAVOUR** message.

Upon successful reservation of resources, it proceeds to create a new contract. Following this, a **PURCHASE\_FLAVOUR** message is sent to the Provider Node's Contract Manager.
