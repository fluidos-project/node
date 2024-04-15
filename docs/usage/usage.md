# Low-level usage

In this section we will instruct you on how you can interact with the FLUIDOS Node using its CRDs.

## Solver Creation

The first step is to create a `Solver` CR. This CR has different fields in its specification that if set correctly can lead to a custom behaviour of the FLUIDOS Node.

Therefore, set the specification of the `Solver` CR as follows:

```yaml
  reserveAndBuy: false
  enstablishPeering: false
```

You can find an example here: [solver.yaml](../../deployments/node/samples/solver-custom.yaml)

Doing so, the FLUIDOS Node, after the creation of the `Solver` CR, will only discover the Peering Candidates adequate for the Solver requests.

Retrieving the `Discovery` CR which contains the SolverID field as the name of the Solver CR, you can see the Peering Candidates discovered by the FLUIDOS Node.

## Reservation

For each Peering Candidate that you want to reserve (temporary action), you need to create a `Reservation` CR.

You can find an example here: [reservation.yaml](../../deployments/node/samples/reservation.yaml)

If you want to postpone the purchase phase, you need to set the `purchase` field to `false` as follows:

```yaml
  reserve: true
  purchase: false
```

Doing so, the FLUIDOS Node will not proceed with the purchase of the Peering Candidate, but you will have a **temporary** reserved Peering Candidate both in the consumer and provider side.

## Purchase

When you want to permanently buy a Peering Candidate, you need to edit its related `Reservation` CR and set the `purchase` field to `true`.

Doing so, the FLUIDOS Node will proceed with the purchase of the Peering Candidate and you will have a contract in your cluster. Information about the contract can be found inside the Status of the `Reservation` CR.

## Peering

After the purchase of the Peering Candidate, the FLUIDOS Node can establish the peering between the consumer and provider. You need to create an `Allocation` CR.

You can find an example here: [allocation.yaml](../../deployments/node/samples/allocation.yaml)

To create this CR, you need to get information from the contract that can be found inside the Status of the `Reservation` CR. More details can be found in the example.

## Conclusion

After all these steps you have reproduced the whole process of the FLUIDOS Node. To be sure that everything went well, we suggest to check at the end the status of the peering by typing the following command:

```bash
liqoctl status peer
```

You should see the peering established with the resources that you have requested.
