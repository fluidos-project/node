#

<p align="center">
<a href="https://www.fluidos.eu/"> <img src="/docs/images/fluidoslogo.png" width="150"/> </a>
<h3 align="center">FLUIDOS Node - Testbed (KIND)</h3>
</p>

## Getting Started

This guide will help you to install a FLUIDOS Node **Testbed**  using KIND (Kubernetes in Docker). This is the easiest way to install FLUIDOS Node on a local machine.

This guide has been made only for testing purposes. If you want to install FLUIDOS Node on a production environment, please follow the [official installation guide](/docs/installation/installation.md)

## What will be installed

This guide will create two different Kubernetes clusters:

* **fluidos-consumer**: This cluster will act as a consumer of the FLUIDOS Node. It will be used to deploy a `solver` example CR which will simulate an Intent resolution request. Through the REAR Protocol it will be able to communicate with the Provider cluster and to receive matching Flavours, reserving the one that best fits the request and purchasing it.

* **fluidos-provider**: This cluster will act as a provider of the FLUIDOS Node. It will offer its own Flavours on the specific request made by the consumer, reserving and selling it.

### Prerequisites

* [Docker](https://docs.docker.com/get-docker/)
* [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
* [KIND](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
* [Helm](https://helm.sh/docs/intro/install/)

### Installation

1) Clone the repository

```sh
git clone https://github.com/fluidos-project/node.git
```

2) Move into the KIND Example folder

```sh
cd examples/kind
```

3) Launch the `setup.sh` script

```sh
. ./setup.sh
```

4) Wait for the script to finish. It will take some minutes.

5) When the script has finished, you can check the status of the pods with the following command:

```sh
kubectl get pods -n fluidos
```

6) You should see 3 pods running on the `fluidos-consumer` cluster and 3 pods running on the `fluidos-provider` cluster:

* `node-local-reaource-manager-<random>`
* `node-rear-manager-<random>`
* `node-rear-controller-<random>`

7) You can also check the status of the generated flavours with the following command:

```sh
k get flavours.nodecore.fluidos.eu -n fluidos  
```

The result should be something like this:

```
NAME                                   PROVIDER ID   TYPE          CPU           MEMORY       OWNER NAME   OWNER DOMAIN   AVAILABLE   AGE
<domain>-k8s-fluidos-<random-suffix>   kc1pttf3vl    k8s-fluidos   4963020133n   26001300Ki   kc1pttf3vl   fluidos.eu     true        168m
<domain>-k8s-fluidos-<random-suffix>   kc1pttf3vl    k8s-fluidos   4954786678n   25966964Ki   kc1pttf3vl   fluidos.eu     true        168m
```

### Usage

Now lets try to deploy a `solver` example CR on the `fluidos-consumer` cluster.

1) Open a new terminal on the repo and move into the `deployments/samples` folder

```sh
cd deployments/samples
```

2) Set the `KUBECONFIG` environment variable to the `fluidos-consumer` cluster

```sh
export KUBECONFIG=../../examples/kind/consumer/config
```

3) Deploy the `solver` CR

```sh
kubectl apply -f solver.yaml
```

4) Check the result of the deployment

```sh
kubectl get solver -n fluidos
```

The result should be something like this:

```
NAMESPACE   NAME            INTENT ID       FIND CANDIDATE   RESERVE AND BUY   PEERING   CANDIDATE PHASE   RESERVING PHASE   PEERING PHASE   STATUS   MESSAGE                           AGE
fluidos     solver-sample   intent-sample   true             true              false     Solved            Solved                            Solved   No need to enstablish a peering   5s
```

5) Other resources have been created, you can check them with the following commands:

```sh
kubectl get flavours.nodecore.fluidos.eu -n fluidos
kubectl get discoveries.advertisement.fluidos.eu -n fluidos
kubectl get reservations.reservation.fluidos.eu -n fluidos
kubectl get contracts.reservation.fluidos.eu -n fluidos
kubectl get peeringcandidates.advertisement.fluidos.eu -n fluidos
kubectl get transactions.reservation.fluidos.eu -n fluidos
```
