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
  name: discovery-solver1
spec:
  selector:
    type: k8s-fluidos
    architecture: arm64
    rangeSelector:
      minCpu: 1
      minMemory: 1
  solverID: solver1
  subscribe: true
```

## Reservation

Here is a `Reservation` sample:

```yaml
apiVersion: reservation.fluidos.eu/v1alpha1
kind: Reservation
metadata:
  name: reservation-solver1
spec:
  buyer:
    domain: topix.fluidos.eu
    ip: 17.3.4.11
    nodeID: 95c0614o1d0
  flavourID: k8s-002
  peeringCandidate:
    name: peeringcandidate-k8s-002
    namespace: default
  purchase: true
  reserve: true
  seller:
    domain: polito.fluidos.eu
    ip: 13.3.5.1
    nodeID: 91cbd32s0q1
```

## Allocation

To be implemented.

## Flavour

Here is a `Flavour` sample:

```yaml
apiVersion: nodecore.fluidos.eu/v1alpha1
kind: Flavour
metadata:
  name: k8s-fluidos-002
  namespace: default
spec:
  characteristics:
    architecture: amd64
    cpu: 4
    memory: 16
  optionalFields:
    availability: true
  owner:
    domain: polito.fluidos.eu
    ip: 13.3.5.1
    nodeID: 91cbd32s0q1
  policy:
    aggregatable:
      maxCount: 5
      minCount: 1
  price:
    amount: "10"
    currency: USD
    period: hourly
  providerID: 05a2a55a-9939-4e94-9587-barlo14
  type: k8s-fluidos
```

## Contract

Here is a `Contract` sample:

```yaml
Name:         contract-k8s-fluidos-002-4o5g
API Version:  reservation.fluidos.eu/v1alpha1
Kind:         Contract
Spec:
  Buyer:
    domain: topix.fluidos.eu
    ip: 17.3.4.11
    nodeID: 95c0614o1d0
  Credentials:
    Cluster ID:
    Cluster Name:
    Endpoint:
    Token:
  Flavour:
    Spec:
      Characteristics:
        Architecture:
        Cpu:                   4
        Ephemeral - Storage:   0
        Gpu:                   0
        Memory:                16
        Persistent - Storage:  0
      Optional Fields:
        Availability:  true
      Owner:
        domain: polito.fluidos.eu
        ip: 13.3.5.1
        nodeID: 91cbd32s0q1
      Policy:
        Aggregatable:
          Max Count:  5
          Min Count:  1
      Price:
        Amount:     10
        Currency:   USD
        Period:     hourly
      Provider ID:  05a2a55a-9939-4e94-9587-barlo14
      Type:         k8s-fluidos
    Status:
      Creation Time: 2023-09-29T10:22:13+02:00
      Expiration Time: 2023-11-29T10:22:13+02:00
      Last Update Time:  2023-09-29T11:26:37+02:00
  Partition:
    Cpu:                0
    Ephemeral Storage:  0
    Gpu:                0
    Memory:             0
    Storage:            0
  Seller:
    domain: polito.fluidos.eu
    ip: 13.3.5.1
    nodeID: 91cbd32s0q1
```

## PeeringCandidate

Here is a `PeeringCandidate` sample:

```yaml
Name:         peeringcandidate-k8s-fluidos-002
API Version:  advertisement.fluidos.eu/v1alpha1
Kind:         PeeringCandidate
Spec:
  Flavour:
    Spec:
      Characteristics:
        Architecture:
        Cpu:                   4
        Ephemeral - Storage:   0
        Gpu:                   0
        Memory:                16
        Persistent - Storage:  0
      Optional Fields:
        Availability:  true
      Owner:
        domain: polito.fluidos.eu
        ip: 13.3.5.1
        nodeID: 91cbd32s0q1
      Policy:
        Aggregatable:
          Max Count:  5
          Min Count:  1
      Price:
        Amount:     10
        Currency:   USD
        Period:     hourly
      Provider ID:  05a2a55a-9939-4e94-9587-barlo14
      Type:         k8s-fluidos-fluidos
    Status:
      Creation Time: 2023-09-29T10:22:13+02:00
      Expiration Time: 2023-11-29T10:22:13+02:00
      Last Update Time:  2023-09-29T11:26:38+02:00
  Reserved:              true
  Solver ID:             solver1
```

## Solver

Here is a `Solver` sample:

```yaml
apiVersion: nodecore.fluidos.eu/v1alpha1
kind: Solver
metadata:
  name: solver1
spec:
  selector:
    type: k8s-fluidos-fluidos
    architecture: arm64
    rangeSelector:
      minCpu: 1
      minMemory: 1
  solverID: solver1
  findCandidate: true
  enstablishPeering: false
```

## Transaction

Here is a `Transaction` sample:

```yaml
Name:         2738d980b9c4c45a172c7a4e18273492-1693646797264948000
API Version:  reservation.fluidos.eu/v1alpha1
Kind:         Transaction
Spec:
  Flavour ID:  k8s-fluidos-002
  Start Time:  2023-09-29T11:26:37+02:00
```
