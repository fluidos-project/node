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
Name:         discovery-solver-sample
Namespace:    fluidos
Labels:       <none>
Annotations:  <none>
API Version:  advertisement.fluidos.eu/v1alpha1
Kind:         Discovery
Metadata:
  Creation Timestamp:  2023-11-16T16:16:44Z
  Generation:          1
  Resource Version:    1593
  UID:                 b084e36b-c9c0-4519-9cbe-7393d3b24cc9
Spec:
  Selector:
    Architecture:  arm64
    Range Selector:
      Max Cpu:      0
      Max Eph:      0
      Max Gpu:      0
      Max Memory:   0
      Max Storage:  0
      Min Cpu:      1
      Min Eph:      0
      Min Gpu:      0
      Min Memory:   1Gi
      Min Storage:  0
    Type:           k8s-fluidos
  Solver ID:        solver-sample
  Subscribe:        false
Status:
  Peering Candidate:
    Name:       peeringcandidate-fluidos.eu-k8s-fluidos-bba29928
    Namespace:  fluidos
  Phase:
    Last Change Time:  2023-11-16T16:16:44Z
    Message:           Discovery Solved: Peering Candidate found
    Phase:             Solved
    Start Time:        2023-11-16T16:16:44Z
Events:                <none>
```

## Reservation

Here is a `Reservation` sample:

```yaml
Name:         reservation-solver-sample
Namespace:    fluidos
Labels:       <none>
Annotations:  <none>
API Version:  reservation.fluidos.eu/v1alpha1
Kind:         Reservation
Metadata:
  Creation Timestamp:  2023-11-16T16:16:44Z
  Generation:          1
  Resource Version:    1606
  UID:                 fb6ad873-7b51-4946-a016-bfdbcff0d2e4
Spec:
  Buyer:
    Domain:   fluidos.eu
    Ip:       172.18.0.2:30000
    Node ID:  jlhfplohpf
  Partition:
    Architecture:         arm64
    Cpu:                  1
    Ephemeral - Storage:  0
    Gpu:                  0
    Memory:               1Gi
    Storage:              0
  Peering Candidate:
    Name:       peeringcandidate-fluidos.eu-k8s-fluidos-bba29928
    Namespace:  fluidos
  Purchase:     true
  Reserve:      true
  Seller:
    Domain:   fluidos.eu
    Ip:       172.18.0.7:30001
    Node ID:  46ltws9per
  Solver ID:  solver-sample
Status:
  Contract:
    Name:       contract-fluidos.eu-k8s-fluidos-bba29928-3c6e
    Namespace:  fluidos
  Phase:
    Last Change Time:  2023-11-16T16:16:45Z
    Message:           Reservation solved
    Phase:             Solved
    Start Time:        2023-11-16T16:16:44Z
  Purchase Phase:      Solved
  Reserve Phase:       Solved
  Transaction ID:      37b53e2ba2b7b6ecb96bd56989baecf2-1700151404800502383
Events:                <none>
```

## Allocation

Here is a `Allocation` sample:

```yaml
Name:         allocation-fluidos.eu-k8s-fluidos-bba29928-3c6e
Namespace:    fluidos
Labels:       <none>
Annotations:  <none>
API Version:  nodecore.fluidos.eu/v1alpha1
Kind:         Allocation
Metadata:
  Creation Timestamp:  2023-11-16T16:16:45Z
  Generation:          1
  Resource Version:    1716
  UID:                 87894f6c-24a9-4389-bd90-a9e9641aa337
Spec:
  Destination:  Local
  Flavour:
    Metadata:
      Name:       fluidos.eu-k8s-fluidos-bba29928
      Namespace:  fluidos
    Spec:
      Characteristics:
        Architecture:          
        Cpu:                   7970838142n
        Ephemeral - Storage:   0
        Gpu:                   0
        Memory:                7879752Ki
        Persistent - Storage:  0
      Optional Fields:
        Availability:  true
        Worker ID:     fluidos-provider-worker
      Owner:
        Domain:   fluidos.eu
        Ip:       172.18.0.7:30001
        Node ID:  46ltws9per
      Policy:
        Aggregatable:
          Max Count:  0
          Min Count:  0
        Partitionable:
          Cpu Min:      0
          Cpu Step:     1
          Memory Min:   0
          Memory Step:  100Mi
      Price:
        Amount:     
        Currency:   
        Period:     
      Provider ID:  46ltws9per
      Type:         k8s-fluidos
    Status:
      Creation Time:     
      Expiration Time:   
      Last Update Time:  
  Intent ID:             solver-sample
  Node Name:             liqo-fluidos-provider
  Partitioned:           true
  Remote Cluster ID:     40585c34-4b93-403f-8008-4b1ecced6f62
  Resources:
    Architecture:          
    Cpu:                   1
    Ephemeral - Storage:   0
    Gpu:                   0
    Memory:                1Gi
    Persistent - Storage:  0
  Type:                    VirtualNode
Status:
  Last Update Time:  2023-11-16T16:16:49Z
  Message:           Outgoing peering ready, Allocation is now Active
  Status:            Active
Events:              <none>
```

## Flavour

Here is a `Flavour` sample:

```yaml
Name:         fluidos.eu-k8s-fluidos-bba29928
Namespace:    fluidos
Labels:       <none>
Annotations:  <none>
API Version:  nodecore.fluidos.eu/v1alpha1
Kind:         Flavour
Metadata:
  Creation Timestamp:  2023-11-16T16:13:52Z
  Generation:          2
  Resource Version:    1534
  UID:                 5ce9f378-014e-4cb4-b173-5a3530d8f78d
Spec:
  Characteristics:
    Architecture:          arm64
    Cpu:                   7970838142n
    Ephemeral - Storage:   0
    Gpu:                   0
    Memory:                7879752Ki
    Persistent - Storage:  0
  Optional Fields:
    Worker ID:  fluidos-provider-worker
  Owner:
    Domain:   fluidos.eu
    Ip:       172.18.0.7:30001
    Node ID:  46ltws9per
  Policy:
    Aggregatable:
      Max Count:  0
      Min Count:  0
    Partitionable:
      Cpu Min:      0
      Cpu Step:     1
      Memory Min:   0
      Memory Step:  100Mi
  Price:
    Amount:     
    Currency:   
    Period:     
  Provider ID:  46ltws9per
  Type:         k8s-fluidos
Events:         <none>
```

## Contract

Here is a `Contract` sample:

```yaml
Name:         contract-fluidos.eu-k8s-fluidos-bba29928-3c6e
Namespace:    fluidos
Labels:       <none>
Annotations:  <none>
API Version:  reservation.fluidos.eu/v1alpha1
Kind:         Contract
Metadata:
  Creation Timestamp:  2023-11-16T16:16:45Z
  Generation:          1
  Resource Version:    1531
  UID:                 baa27f94-ebd8-4e7b-98fe-3e7d5b09d4bb
Spec:
  Buyer:
    Domain:          fluidos.eu
    Ip:              172.18.0.2:30000
    Node ID:         jlhfplohpf
  Buyer Cluster ID:  14461b0e-446d-4b05-b1f8-9ddb6765ac02
  Expiration Time:   2024-11-15T16:16:45Z
  Flavour:
    API Version:  nodecore.fluidos.eu/v1alpha1
    Kind:         Flavour
    Metadata:
      Name:       fluidos.eu-k8s-fluidos-bba29928
      Namespace:  fluidos
    Spec:
      Characteristics:
        Architecture:          arm64
        Cpu:                   7970838142n
        Ephemeral - Storage:   0
        Gpu:                   0
        Memory:                7879752Ki
        Persistent - Storage:  0
      Optional Fields:
        Availability:  true
        Worker ID:     fluidos-provider-worker
      Owner:
        Domain:   fluidos.eu
        Ip:       172.18.0.7:30001
        Node ID:  46ltws9per
      Policy:
        Aggregatable:
          Max Count:  0
          Min Count:  0
        Partitionable:
          Cpu Min:      0
          Cpu Step:     1
          Memory Min:   0
          Memory Step:  100Mi
      Price:
        Amount:     
        Currency:   
        Period:     
      Provider ID:  46ltws9per
      Type:         k8s-fluidos
    Status:
      Creation Time:     
      Expiration Time:   
      Last Update Time:  
  Partition:
    Architecture:         
    Cpu:                  1
    Ephemeral - Storage:  0
    Gpu:                  0
    Memory:               1Gi
    Storage:              0
  Seller:
    Domain:   fluidos.eu
    Ip:       172.18.0.7:30001
    Node ID:  46ltws9per
  Seller Credentials:
    Cluster ID:    40585c34-4b93-403f-8008-4b1ecced6f62
    Cluster Name:  fluidos-provider
    Endpoint:      https://172.18.0.6:31780
    Token:         0959ee7a6290b51cb223f971e30455043a1af01b383b5dc8cd13650a4d61e9b6184c524fc56699d427b37044fa1e3b05179ecf19f807aa67d26fcaa1cebe4d68
  Transaction ID:  37b53e2ba2b7b6ecb96bd56989baecf2-1700151404800502383
Events:            <none>
```

## PeeringCandidate

Here is a `PeeringCandidate` sample:

```yaml
Name:         peeringcandidate-fluidos.eu-k8s-fluidos-bba29928
Namespace:    fluidos
Labels:       <none>
Annotations:  <none>
API Version:  advertisement.fluidos.eu/v1alpha1
Kind:         PeeringCandidate
Metadata:
  Creation Timestamp:  2023-11-16T16:16:44Z
  Generation:          1
  Resource Version:    1592
  UID:                 030c441f-a17e-4778-9515-9cd4293f656a
Spec:
  Flavour:
    Metadata:
      Name:       fluidos.eu-k8s-fluidos-bba29928
      Namespace:  fluidos
    Spec:
      Characteristics:
        Architecture:          
        Cpu:                   7970838142n
        Ephemeral - Storage:   0
        Gpu:                   0
        Memory:                7879752Ki
        Persistent - Storage:  0
      Optional Fields:
        Availability:  true
        Worker ID:     fluidos-provider-worker
      Owner:
        Domain:   fluidos.eu
        Ip:       172.18.0.7:30001
        Node ID:  46ltws9per
      Policy:
        Aggregatable:
          Max Count:  0
          Min Count:  0
        Partitionable:
          Cpu Min:      0
          Cpu Step:     1
          Memory Min:   0
          Memory Step:  100Mi
      Price:
        Amount:     
        Currency:   
        Period:     
      Provider ID:  46ltws9per
      Type:         k8s-fluidos
    Status:
      Creation Time:     
      Expiration Time:   
      Last Update Time:  
  Reserved:              true
  Solver ID:             solver-sample
Events:                  <none>
```

## Solver

Here is a `Solver` sample:

```yaml
Name:         solver-sample
Namespace:    fluidos
Labels:       <none>
Annotations:  <none>
API Version:  nodecore.fluidos.eu/v1alpha1
Kind:         Solver
Metadata:
  Creation Timestamp:  2023-11-16T16:16:44Z
  Generation:          1
  Resource Version:    1717
  UID:                 392ae554-69e4-4639-90ef-34d8f3a83aef
Spec:
  Enstablish Peering:  true
  Find Candidate:      true
  Intent ID:           intent-sample
  Reserve And Buy:     true
  Selector:
    Architecture:  arm64
    Range Selector:
      Min Cpu:     1000m
      Min Memory:  1Gi
    Type:          k8s-fluidos
Status:
  Allocation:
    Name:       allocation-fluidos.eu-k8s-fluidos-bba29928-3c6e
    Namespace:  fluidos
  Contract:
    Name:       contract-fluidos.eu-k8s-fluidos-bba29928-3c6e
    Namespace:  fluidos
  Credentials:
    Cluster ID:     40585c34-4b93-403f-8008-4b1ecced6f62
    Cluster Name:   fluidos-provider
    Endpoint:       https://172.18.0.6:31780
    Token:          0959ee7a6290b51cb223f971e30455043a1af01b383b5dc8cd13650a4d61e9b6184c524fc56699d427b37044fa1e3b05179ecf19f807aa67d26fcaa1cebe4d68
  Discovery Phase:  Solved
  Find Candidate:   Solved
  Peering:          Solved
  Peering Candidate:
    Name:             peeringcandidate-fluidos.eu-k8s-fluidos-bba29928
    Namespace:        fluidos
  Reservation Phase:  Solved
  Reserve And Buy:    Solved
  Solver Phase:
    End Time:          2023-11-16T16:16:49Z
    Last Change Time:  2023-11-16T16:16:49Z
    Message:           Solver has enstablished a peering
    Phase:             Solved
Events:                <none>
```

## Transaction

Here is a `Transaction` sample:

```yaml
Name:         37b53e2ba2b7b6ecb96bd56989baecf2-1700151404800502383
Namespace:    fluidos
Labels:       <none>
Annotations:  <none>
API Version:  reservation.fluidos.eu/v1alpha1
Kind:         Transaction
Metadata:
  Creation Timestamp:  2023-11-16T16:16:44Z
  Generation:          1
  Resource Version:    1600
  UID:                 a1d73148-0795-4061-80cf-197e6a83379c
Spec:
  Buyer:
    Domain:    fluidos.eu
    Ip:        172.18.0.2:30000
    Node ID:   jlhfplohpf
  Cluster ID:  14461b0e-446d-4b05-b1f8-9ddb6765ac02
  Flavour ID:  fluidos.eu-k8s-fluidos-bba29928
  Partition:
    Architecture:         
    Cpu:                  1
    Ephemeral - Storage:  0
    Gpu:                  0
    Memory:               1Gi
    Storage:              0
  Start Time:             2023-11-16T16:16:44Z
Events:                   <none>
```
