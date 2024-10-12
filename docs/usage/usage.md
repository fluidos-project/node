# Low-level usage

In this section we will instruct you on how you can interact with the FLUIDOS Node using its CRDs.

## Solver Creation

The first step is to create a `Solver` CR. This CR has different fields in its specification that if set correctly can lead to a custom behavior of the FLUIDOS Node.

Therefore, set the specification of the `Solver` CR as follows:

```yaml
  reserveAndBuy: false
  establishPeering: false
```

You can find an example here: [solver.yaml](../../deployments/node/samples/solver-custom.yaml)

Doing so, the FLUIDOS Node, after the creation of the `Solver` CR, will only discover the Peering Candidates adequate for the Solver requests.

Retrieving the `Discovery` CR which contains the SolverID field as the name of the Solver CR, you can see the Peering Candidates discovered by the FLUIDOS Node.

### Selector

In the `Solver` CR you can also set a `Selector` field. This field specifies the requirements that the Peering Candidates must satisfy.

In particular, you have two fields:

- `flavorType`: that specifies the flavor type of the Peering Candidate. The possible values are:
  - `K8Slice`: for a Peering Candidate that has a Kubernetes Slice flavor.
  - `VM`: for a Peering Candidate that has a VM flavor.
  - `Service`: for a Peering Candidate that has a Service flavor.
  - `Sensor`: for a Peering Candidate that has a Sensor flavor.
- `filters`: that specifies the filters that the Peering Candidates must satisfy. These filters depends on the flavor type of the Peering Candidate that you want to discover.

### Filter types

Each filter inside a `Selector` field can be of different types. The possible types are:

#### ResourceQuantityFilter

The `ResourceQuantityFilter` is a specific type of filter used by some specific filters in specific selectors. It can be of two types and it is specified by the `name` field. The possible values are:

- `Match`: that specifies the exact value that the Peering Candidate must have.
- `Range`: that specifies a range of values that the Peering Candidate must have.

The other field is `data` that specifies the effective value of the filter and its structure depends on the type of the filter.

- `Match` filter: Its structure is as follows:

```yaml
  name: Match
  data:
    value: "1000m"
```

- `Range` filter: Its structure is as follows:

```yaml
  name: Range
  data:
    min: "1Gi"
    max: "10Gi"
```

The **Range** filter can have only the `min` field, the `max` field or both.

#### StringFilter

The `StringFilter` is a specific type of filter used by some specific filters in specific selectors. It can be of two types and it is specified by the `name` field. The possible values are:

- `Match`: that specifies the exact value that the Peering Candidate must have.
- `Range`: that specifies a range of values (array) that the Peering Candidate can have (at least one of them).

The other field is `data` that specifies the effective value of the filter and its structure depends on the type of the filter.

- `Match` filter: Its structure is as follows:

```yaml
  name: Match
  data:
    value: "cateogory-example"
```

- `Range` filter: Its structure is as follows:

```yaml
  name: Range
  data:
    values:
      - "tag1"
      - "tag2"
```

### Selector types

The `Selector` field can be of different types. The possible types are:

#### K8Slice Filters

K8Slice filters are the filters that you can set for a Peering Candidate that has a Kubernetes Slice flavor. They are specifically fields:

- `CpuFilter`: that specifies the CPU that the Peering Candidate must have. It is a `ResourceQuantityFilter` and can be both a `Match` or a `Range` filter type. (Optional)
- `MemoryFilter`: that specifies the Memory that the Peering Candidate must have. It is a `ResourceQuantityFilter` and can be both a `Match` or a `Range` filter type. (Optional)
- `PodsFilter`: that specifies the Architecture that the Peering Candidate must have. It is a `ResourceQuantityFilter` and can be both a `Match` or a `Range` filter type. (Optional)
- `StorageFilter`: that specifies the Storage that the Peering Candidate must have. It is a `ResourceQuantityFilter` and can be both a `Match` or a `Range` filter type. (Optional)

A structure example of a `K8Slice` filter is as follows:

```yaml
  cpuFilter:
    name: Match
    value: "1000m"
  memoryFilter:
    name: Range
    min: "1Gi"
    max: "10Gi"
  podsFilter:
    name: Range
    max: 10
```

#### VM Filters

*Not yet implemented.*

#### Service Filters

Service filters are the filters that you can set for a Peering Candidate that has a Service flavor. They are specifically fields:

- `CategoryFilter`: that specifies the Category that the Peering Candidate must have. It is a `StringFilter` and can be both a `Match` or a `Range` filter type. (Optional)
- `TagsFilter`: that specifies the Tags that the Peering Candidate must have. It is a `StringFilter` and can be both a `Match` or a `Range` filter type. (Optional)

A structure example of a `Service` filter is as follows:

```yaml
  categoryFilter:
    name: Match
    value: "category-example"
  tagsFilter:
    name: Range
    values:
      - "tag1"
      - "tag2"
```

#### Sensor Filters

*Not yet implemented.*

## Reservation

For each Peering Candidate that you want to reserve (temporary action), you need to create a `Reservation` CR.

You can find an example here: [reservation.yaml](../../deployments/node/samples/reservation.yaml)

If you want to postpone the purchase phase, you need to set the `purchase` field to `false` as follows:

```yaml
  reserve: true
  purchase: false
```

Doing so, the FLUIDOS Node will not proceed with the purchase of the Peering Candidate, but you will have a **temporary** reserved Peering Candidate both in the consumer and provider side.

## Configuration

In the `Reservation` CR you can define an optional field called `configuration`. This field is used to specify the configuration of the Peering Candidate that you want to reserve.

The structure of the `configuration` field depends on the flavor type of the Peering Candidate that you want to reserve. In fact, there are two fields that you have to set:

- `type`: that specifies the flavor type of the Configuration. The possible values are:
  - `K8Slice`: for a Peering Candidate that has a Kubernetes Slice flavor.
  - `VM`: for a Peering Candidate that has a VM flavor.
  - `Service`: for a Peering Candidate that has a Service flavor.
  - `Sensor`: for a Peering Candidate that has a Sensor flavor.
- `data`: that specifies the specific configuration parameters of the Peering Candidate, related to the type of the Configuration.

### Configuration types

The `configuration` field can be of different types. The possible types are:

#### K8Slice Configuration

K8Slice configuration is the configuration that you can set for a Peering Candidate that has a Kubernetes Slice flavor. The configuration of the K8Slice leads to the creation of a partition of the Peering Candidate. The fields that you have set are:

- `cpu`: that specifies the CPU of the Partition
- `memory`: that specifies the Memory of the Partition
- `pods`: that specifies the Pods of the Partition
- `gpu`: that specifies the GPU of the Partition (Optional, not yet implemented)
- `storage`: hat specifies the Storage of the Partition (Optional)

A structure example of a `K8Slice` configuration is as follows:

```yaml
  type: K8Slice
  data:
    cpu: "1000m"
    memory: "1Gi"
    pods: "10"
```

#### Service Configuration

Service configuration is the configuration that you can set for a Peering Candidate that has a Service flavor. The configuration of the Service leads to configuration of the service that will be provided. The fields that you have set are:

- `hostingPolicy`: that specifies the Hosting Policy chosen for the Service. The set of possibile values are previously defined by the Flavor chosen to be reserved and configured, but they generally are:
  - `Provider`: for a Service that will be hosted by the Provider.
  - `Consumer`: for a Service that will be hosted by the Consumer.
  - `Shared`: for a Service that will be hosted by both the Provider and the Consumer, effective scheduling will be in charge of the scheduler of the provider cluster.

## Purchase

When you want to permanently buy a Peering Candidate, you need to edit its related `Reservation` CR and set the `purchase` field to `true`.

Doing so, the FLUIDOS Node will proceed with the purchase of the Peering Candidate and you will have a contract in your cluster. Information about the contract can be found inside the Status of the `Reservation` CR.

## Allocation

After the purchase of the Peering Candidate, the FLUIDOS Node can establish the peering between the consumer and provider and allocate the negotiated resources. You need to create an `Allocation` CR.

You can find an example here: [allocation.yaml](../../deployments/node/samples/allocation.yaml)

To create this CR, you need to get information from the contract that can be found inside the Status of the `Reservation` CR. More details can be found in the example.

## Conclusion

After all these steps you have reproduced the whole process of the FLUIDOS Node. To be sure that everything went well, we suggest to check at the end the status of the peering by typing the following command:

```bash
liqoctl status peer
```

You should see the peering established with the resources that you have requested.
