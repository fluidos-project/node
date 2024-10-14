# Custom Resources

The following custom resources have been developed for the FLUIDOS Node:

- [**Discovery**](./customresources.md#discovery)
- [**Reservation**](./customresources.md#reservation)
- [**Allocation**](./customresources.md#allocation)
- [**Flavour**](./customresources.md#flavour)
- [**Contract**](./customresources.md#contract)
- [**PeeringCandidate**](./customresources.md#peeringcandidate)
- [**Solver**](./customresources.md#solver)
- [**Transaction**](./customresources.md#transaction)

## Discovery

Here is a `Discovery` sample:

```yaml
apiVersion: advertisement.fluidos.eu/v1alpha1
kind: Discovery
metadata:
  creationTimestamp: "2023-11-16T16:16:44Z"
  generation: 1
  name: discovery-solver-sample
  namespace: fluidos
  resourceVersion: "1593"
  uid: b084e36b-c9c0-4519-9cbe-7393d3b24cc9
spec:
  selector:
    architecture: arm64
    rangeSelector:
      MaxCpu: "0"
      MaxEph: "0"
      MaxGpu: "0"
      MaxMemory: "0"
      MaxStorage: "0"
      minCpu: "1"
      minEph: "0"
      minGpu: "0"
      minMemory: 1Gi
      minStorage: "0"
    type: k8s-fluidos
  solverID: solver-sample
  subscribe: false
status:
  peeringCandidate:
    name: peeringcandidate-fluidos.eu-k8s-fluidos-bba29928
    namespace: fluidos
  phase:
    lastChangeTime: "2023-11-16T16:16:44Z"
    message: 'Discovery Solved: Peering Candidate found'
    phase: Solved
    startTime: "2023-11-16T16:16:44Z"
```

## Reservation

Here is a `Reservation` sample:

```yaml
apiVersion: reservation.fluidos.eu/v1alpha1
kind: Reservation
metadata:
  creationTimestamp: "2023-11-16T16:16:44Z"
  generation: 1
  name: reservation-solver-sample
  namespace: fluidos
  resourceVersion: "1606"
  uid: fb6ad873-7b51-4946-a016-bfdbcff0d2e4
spec:
  buyer:
    domain: fluidos.eu
    ip: 172.18.0.2:30000
    nodeID: jlhfplohpf
  partition:
    architecture: arm64
    cpu: "1"
    ephemeral-storage: "0"
    gpu: "0"
    memory: 1Gi
    storage: "0"
  peeringCandidate:
    name: peeringcandidate-fluidos.eu-k8s-fluidos-bba29928
    namespace: fluidos
  purchase: true
  reserve: true
  seller:
    domain: fluidos.eu
    ip: 172.18.0.7:30001
    nodeID: 46ltws9per
  solverID: solver-sample
status:
  contract:
    name: contract-fluidos.eu-k8s-fluidos-bba29928-3c6e
    namespace: fluidos
  phase:
    lastChangeTime: "2023-11-16T16:16:45Z"
    message: Reservation solved
    phase: Solved
    startTime: "2023-11-16T16:16:44Z"
  purchasePhase: Solved
  reservePhase: Solved
  transactionID: 37b53e2ba2b7b6ecb96bd56989baecf2-1700151404800502383
```

## Allocation

Here is a `Allocation` sample:

```yaml
apiVersion: nodecore.fluidos.eu/v1alpha1
kind: Allocation
metadata:
  creationTimestamp: "2023-11-16T16:16:45Z"
  generation: 1
  name: allocation-fluidos.eu-k8s-fluidos-bba29928-3c6e
  namespace: fluidos
  resourceVersion: "1716"
  uid: 87894f6c-24a9-4389-bd90-a9e9641aa337
spec:
  destination: Local
  flavour:
    metadata:
      name: fluidos.eu-k8s-fluidos-bba29928
      namespace: fluidos
    spec:
      characteristics:
        architecture: ""
        cpu: 7970838142n
        ephemeral-storage: "0"
        gpu: "0"
        memory: 7879752Ki
        persistent-storage: "0"
      optionalFields:
        availability: true
        workerID: fluidos-provider-worker
      owner:
        domain: fluidos.eu
        ip: 172.18.0.7:30001
        nodeID: 46ltws9per
      policy:
        aggregatable:
          maxCount: 0
          minCount: 0
        partitionable:
          cpuMin: "0"
          cpuStep: "1"
          memoryMin: "0"
          memoryStep: 100Mi
      price:
        amount: ""
        currency: ""
        period: ""
      providerID: 46ltws9per
      type: k8s-fluidos
    status:
      creationTime: ""
      expirationTime: ""
      lastUpdateTime: ""
  intentID: solver-sample
  nodeName: liqo-fluidos-provider
  partitioned: true
  remoteClusterID: 40585c34-4b93-403f-8008-4b1ecced6f62
  resources:
    architecture: ""
    cpu: "1"
    ephemeral-storage: "0"
    gpu: "0"
    memory: 1Gi
    persistent-storage: "0"
  type: VirtualNode
status:
  lastUpdateTime: "2023-11-16T16:16:49Z"
  message: Outgoing peering ready, Allocation is now Active
  status: Active
```

## Flavour

Here is a `Flavour` sample:

```yaml
apiVersion: nodecore.fluidos.eu/v1alpha1
kind: Flavour
metadata:
  creationTimestamp: "2023-11-16T16:13:52Z"
  generation: 2
  name: fluidos.eu-k8s-fluidos-bba29928
  namespace: fluidos
  resourceVersion: "1534"
  uid: 5ce9f378-014e-4cb4-b173-5a3530d8f78d
spec:
  characteristics:
    architecture: arm64
    cpu: 7970838142n
    ephemeral-storage: "0"
    gpu: "0"
    memory: 7879752Ki
    persistent-storage: "0"
  optionalFields:
    workerID: fluidos-provider-worker
  owner:
    domain: fluidos.eu
    ip: 172.18.0.7:30001
    nodeID: 46ltws9per
  policy:
    aggregatable:
      maxCount: 0
      minCount: 0
    partitionable:
      cpuMin: "0"
      cpuStep: "1"
      memoryMin: "0"
      memoryStep: 100Mi
  price:
    amount: ""
    currency: ""
    period: ""
  providerID: 46ltws9per
  type: k8s-fluidos
```

## Contract

Here is a `Contract` sample:

```yaml
apiVersion: reservation.fluidos.eu/v1alpha1
kind: Contract
metadata:
  creationTimestamp: "2023-11-16T16:16:45Z"
  generation: 1
  name: contract-fluidos.eu-k8s-fluidos-bba29928-3c6e
  namespace: fluidos
  resourceVersion: "1531"
  uid: baa27f94-ebd8-4e7b-98fe-3e7d5b09d4bb
spec:
  buyer:
    domain: fluidos.eu
    ip: 172.18.0.2:30000
    nodeID: jlhfplohpf
  buyerClusterID: 14461b0e-446d-4b05-b1f8-9ddb6765ac02
  expirationTime: "2024-11-15T16:16:45Z"
  flavour:
    apiVersion: nodecore.fluidos.eu/v1alpha1
    kind: Flavour
    metadata:
      name: fluidos.eu-k8s-fluidos-bba29928
      namespace: fluidos
    spec:
      characteristics:
        architecture: arm64
        cpu: 7970838142n
        ephemeral-storage: "0"
        gpu: "0"
        memory: 7879752Ki
        persistent-storage: "0"
      optionalFields:
        availability: true
        workerID: fluidos-provider-worker
      owner:
        domain: fluidos.eu
        ip: 172.18.0.7:30001
        nodeID: 46ltws9per
      policy:
        aggregatable:
          maxCount: 0
          minCount: 0
        partitionable:
          cpuMin: "0"
          cpuStep: "1"
          memoryMin: "0"
          memoryStep: 100Mi
      price:
        amount: ""
        currency: ""
        period: ""
      providerID: 46ltws9per
      type: k8s-fluidos
    status:
      creationTime: ""
      expirationTime: ""
      lastUpdateTime: ""
  partition:
    architecture: ""
    cpu: "1"
    ephemeral-storage: "0"
    gpu: "0"
    memory: 1Gi
    storage: "0"
  seller:
    domain: fluidos.eu
    ip: 172.18.0.7:30001
    nodeID: 46ltws9per
  sellerCredentials:
    clusterID: 40585c34-4b93-403f-8008-4b1ecced6f62
    clusterName: fluidos-provider
    endpoint: https://172.18.0.6:31780
    token: 0959ee7a6290b51cb223f971e30455043a1af01b383b5dc8cd13650a4d61e9b6184c524fc56699d427b37044fa1e3b05179ecf19f807aa67d26fcaa1cebe4d68
  transactionID: 37b53e2ba2b7b6ecb96bd56989baecf2-1700151404800502383
```

## PeeringCandidate

Here is a `PeeringCandidate` sample:

```yaml
apiVersion: advertisement.fluidos.eu/v1alpha1
kind: PeeringCandidate
metadata:
  creationTimestamp: "2023-11-16T16:16:44Z"
  generation: 1
  name: peeringcandidate-fluidos.eu-k8s-fluidos-bba29928
  namespace: fluidos
  resourceVersion: "1592"
  uid: 030c441f-a17e-4778-9515-9cd4293f656a
spec:
  flavour:
    metadata:
      name: fluidos.eu-k8s-fluidos-bba29928
      namespace: fluidos
    spec:
      characteristics:
        architecture: ""
        cpu: 7970838142n
        ephemeral-storage: "0"
        gpu: "0"
        memory: 7879752Ki
        persistent-storage: "0"
      optionalFields:
        availability: true
        workerID: fluidos-provider-worker
      owner:
        domain: fluidos.eu
        ip: 172.18.0.7:30001
        nodeID: 46ltws9per
      policy:
        aggregatable:
          maxCount: 0
          minCount: 0
        partitionable:
          cpuMin: "0"
          cpuStep: "1"
          memoryMin: "0"
          memoryStep: 100Mi
      price:
        amount: ""
        currency: ""
        period: ""
      providerID: 46ltws9per
      type: k8s-fluidos
    status:
      creationTime: ""
      expirationTime: ""
      lastUpdateTime: ""
  reserved: true
  solverID: solver-sample
```

## Solver

Here is a `Solver` sample:

```yaml
apiVersion: nodecore.fluidos.eu/v1alpha1
kind: Solver
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"nodecore.fluidos.eu/v1alpha1","kind":"Solver","metadata":{"annotations":{},"name":"solver-sample","namespace":"fluidos"},"spec":{"enstablishPeering":true,"findCandidate":true,"intentID":"intent-sample","reserveAndBuy":true,"selector":{"architecture":"arm64","rangeSelector":{"minCpu":"1000m","minMemory":"1Gi"},"type":"k8s-fluidos"}}}
  creationTimestamp: "2023-11-16T16:16:44Z"
  generation: 1
  name: solver-sample
  namespace: fluidos
  resourceVersion: "1717"
  uid: 392ae554-69e4-4639-90ef-34d8f3a83aef
spec:
  enstablishPeering: true
  findCandidate: true
  intentID: intent-sample
  reserveAndBuy: true
  selector:
    architecture: arm64
    rangeSelector:
      minCpu: 1000m
      minMemory: 1Gi
    type: k8s-fluidos
status:
  allocation:
    name: allocation-fluidos.eu-k8s-fluidos-bba29928-3c6e
    namespace: fluidos
  contract:
    name: contract-fluidos.eu-k8s-fluidos-bba29928-3c6e
    namespace: fluidos
  credentials:
    clusterID: 40585c34-4b93-403f-8008-4b1ecced6f62
    clusterName: fluidos-provider
    endpoint: https://172.18.0.6:31780
    token: 0959ee7a6290b51cb223f971e30455043a1af01b383b5dc8cd13650a4d61e9b6184c524fc56699d427b37044fa1e3b05179ecf19f807aa67d26fcaa1cebe4d68
  discoveryPhase: Solved
  findCandidate: Solved
  peering: Solved
  peeringCandidate:
    name: peeringcandidate-fluidos.eu-k8s-fluidos-bba29928
    namespace: fluidos
  reservationPhase: Solved
  reserveAndBuy: Solved
  solverPhase:
    endTime: "2023-11-16T16:16:49Z"
    lastChangeTime: "2023-11-16T16:16:49Z"
    message: New active peering, triggered by Solver
    phase: Solved
```

## Transaction

Here is a `Transaction` sample:

```yaml
apiVersion: reservation.fluidos.eu/v1alpha1
kind: Transaction
metadata:
  creationTimestamp: "2023-11-16T16:16:44Z"
  generation: 1
  name: 37b53e2ba2b7b6ecb96bd56989baecf2-1700151404800502383
  namespace: fluidos
  resourceVersion: "1600"
  uid: a1d73148-0795-4061-80cf-197e6a83379c
spec:
  buyer:
    domain: fluidos.eu
    ip: 172.18.0.2:30000
    nodeID: jlhfplohpf
  clusterID: 14461b0e-446d-4b05-b1f8-9ddb6765ac02
  flavourID: fluidos.eu-k8s-fluidos-bba29928
  partition:
    architecture: ""
    cpu: "1"
    ephemeral-storage: "0"
    gpu: "0"
    memory: 1Gi
    storage: "0"
  startTime: "2023-11-16T16:16:44Z"
```
