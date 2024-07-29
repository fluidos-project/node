# node

![Version: 0.0.1](https://img.shields.io/badge/Version-0.0.1-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.0.1](https://img.shields.io/badge/AppVersion-0.0.1-informational?style=flat-square)

A Helm chart for Fluidos Node

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| common.affinity | object | `{}` | Affinity for all fluidos-node pods |
| common.extraArgs | list | `[]` | Extra arguments for all fluidos-node pods |
| common.nodeSelector | object | `{}` | NodeSelector for all fluidos-node pods |
| common.tolerations | list | `[]` | Tolerations for all fluidos-node pods |
| localResourceManager.config.enableAutoDiscovery | bool | `true` | Enable the auto-discovery of the resources. |
| localResourceManager.config.flavor.cpuMin | string | `"0"` | The minimum number of CPUs that can be requested to purchase a flavor. |
| localResourceManager.config.flavor.cpuStep | string | `"1000m"` | The CPU step that must be respected when requesting a flavor through a Flavor Selector. |
| localResourceManager.config.flavor.memoryMin | string | `"0"` | The minimum amount of memory that can be requested to purchase a flavor. |
| localResourceManager.config.flavor.memoryStep | string | `"100Mi"` | The memory step that must be respected when requesting a flavor through a Flavor Selector. |
| localResourceManager.config.nodeResourceLabel | string | `"node-role.fluidos.eu/resources"` | Label used to identify the nodes from which resources are collected. |
| localResourceManager.config.resourceType | string | `"k8s-fluidos"` | This flag defines the resource type of the generated flavors. |
| localResourceManager.imageName | string | `"ghcr.io/fluidos-project/local-resource-manager"` |  |
| localResourceManager.pod.annotations | object | `{}` | Annotations for the local-resource-manager pod. |
| localResourceManager.pod.extraArgs | list | `[]` | Extra arguments for the local-resource-manager pod. |
| localResourceManager.pod.labels | object | `{}` | Labels for the local-resource-manager pod. |
| localResourceManager.pod.resources | object | `{"limits":{},"requests":{}}` | Resource requests and limits (https://kubernetes.io/docs/user-guide/compute-resources/) for the local-resource-manager pod. |
| localResourceManager.replicas | int | `1` | The number of REAR Controller, which can be increased for active/passive high availability. |
| networkManager.configMaps.nodeIdentity.domain | string | `""` | The domain name of the FLUIDOS closed domani: It represents for instance the Enterprise and it is used to generate the FQDN of the owned FLUIDOS Nodes |
| networkManager.configMaps.nodeIdentity.ip | string | `nil` | The IP address of the FLUIDOS Node. It can be public or private, depending on the network configuration and it corresponds to the IP address to reach the Network Manager from the outside of the cluster. |
| networkManager.configMaps.nodeIdentity.name | string | `"fluidos-network-manager-identity"` | The name of the ConfigMap containing the FLUIDOS Node identity info. |
| networkManager.configMaps.nodeIdentity.nodeID | string | `nil` | The NodeID is a UUID that identifies the FLUIDOS Node. It is used to generate the FQDN of the owned FLUIDOS Nodes and it is unique in the FLUIDOS closed domain |
| networkManager.configMaps.providers.default | string | `nil` | The IP List of SuperNodes separated by commas. |
| networkManager.configMaps.providers.local | string | `""` | The IP List of Local knwon FLUIDOS Nodes separated by commas. |
| networkManager.configMaps.providers.name | string | `"fluidos-network-manager-config"` | The name of the ConfigMap containing the list of the FLUIDOS Providers and the default FLUIDOS Provider (SuperNode or Catalogue). |
| networkManager.configMaps.providers.remote | string | `nil` | The IP List of Remote known FLUIDOS Nodes separated by commas. |
| networkManager.imageName | string | `"ghcr.io/fluidos-project/network-manager"` |  |
| networkManager.pod.annotations | object | `{}` | Annotations for the network-manager pod. |
| networkManager.pod.extraArgs | list | `[]` | Extra arguments for the network-manager pod. |
| networkManager.pod.labels | object | `{}` | Labels for the network-manager pod. |
| networkManager.pod.resources | object | `{"limits":{},"requests":{}}` | Resource requests and limits (https://kubernetes.io/docs/user-guide/compute-resources/) for the network-manager pod. |
| networkManager.replicas | int | `1` | The number of Network Manager, which can be increased for active/passive high availability. |
| provider | string | `"your-provider"` |  |
| pullPolicy | string | `"IfNotPresent"` | The pullPolicy for fluidos-node pods. |
| rearController.imageName | string | `"ghcr.io/fluidos-project/rear-controller"` |  |
| rearController.pod.annotations | object | `{}` | Annotations for the rear-controller pod. |
| rearController.pod.extraArgs | list | `[]` | Extra arguments for the rear-controller pod. |
| rearController.pod.labels | object | `{}` | Labels for the rear-controller pod. |
| rearController.pod.resources | object | `{"limits":{},"requests":{}}` | Resource requests and limits (https://kubernetes.io/docs/user-guide/compute-resources/) for the rear-controller pod. |
| rearController.replicas | int | `1` | The number of REAR Controller, which can be increased for active/passive high availability. |
| rearController.service.gateway.annotations | object | `{}` | Annotations for the REAR gateway service. |
| rearController.service.gateway.labels | object | `{}` | Labels for the REAR gateway service. |
| rearController.service.gateway.loadBalancer | object | `{"ip":""}` | Options valid if service type is LoadBalancer. |
| rearController.service.gateway.loadBalancer.ip | string | `""` | Override the IP here if service type is LoadBalancer and you want to use a specific IP address, e.g., because you want a static LB. |
| rearController.service.gateway.name | string | `"gateway"` |  |
| rearController.service.gateway.nodePort | object | `{"port":""}` | Options valid if service type is NodePort. |
| rearController.service.gateway.nodePort.port | string | `""` | Force the port used by the NodePort service. |
| rearController.service.gateway.port | int | `3004` | The port used by the rear-controller to expose the REAR Gateway. |
| rearController.service.gateway.targetPort | int | `3004` | The target port used by the REAR Gateway service. |
| rearController.service.gateway.type | string | `"NodePort"` | Kubernetes service to be used to expose the REAR gateway. |
| rearController.service.grpc.annotations | object | `{}` | Annotations for the gRPC service. |
| rearController.service.grpc.labels | object | `{}` | Labels for the gRPC service. |
| rearController.service.grpc.name | string | `"grpc"` |  |
| rearController.service.grpc.port | int | `2710` | The gRPC port used by Liqo to connect with the Gateway of the rear-controller to obtain the Contract resources for a given consumer ClusterID. |
| rearController.service.grpc.targetPort | int | `2710` | The target port used by the gRPC service. |
| rearController.service.grpc.type | string | `"ClusterIP"` | Kubernetes service used to expose the gRPC Server to liqo. |
| rearManager.imageName | string | `"ghcr.io/fluidos-project/rear-manager"` |  |
| rearManager.pod.annotations | object | `{}` | Annotations for the rear-manager pod. |
| rearManager.pod.extraArgs | list | `[]` | Extra arguments for the rear-manager pod. |
| rearManager.pod.labels | object | `{}` | Labels for the rear-manager pod. |
| rearManager.pod.resources | object | `{"limits":{},"requests":{}}` | Resource requests and limits (https://kubernetes.io/docs/user-guide/compute-resources/) for the rear-manager pod. |
| rearManager.replicas | int | `1` | The number of REAR Manager, which can be increased for active/passive high availability. |
| tag | string | `""` | Images' tag to select a development version of fluidos-node instead of a release |
| webhook.deployment | object | `{"certsMount":"/tmp/k8s-webhook-server/serving-certs/"}` | Configuration for the webhook server. |
| webhook.deployment.certsMount | string | `"/tmp/k8s-webhook-server/serving-certs/"` | The mount path for the webhook certificates. |
| webhook.enabled | bool | `true` | Enable the webhook server for the local-resource-manager. |
| webhook.issuer | string | `"self-signed"` | Configuration for the webhook server. |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.11.0](https://github.com/norwoodj/helm-docs/releases/v1.11.0)
