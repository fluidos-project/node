# Controllers

In the following, the controllers developed for the FLUIDOS Node are described. To see the different objects, see [**Custom Resources**](./customresources.md#custom-resources) part.

## Solver Controller (`solver_controller.go`)

The Solver controller, tasked with reconciliation on the `Solver` object, continuously monitors and manages its state to ensure alignment with the desired configuration. It follows the following steps:

1. When there is a new Solver object, it firstly checks if the `Solver` has expired or failed (if so, it marks the Solver as `Timed Out`).
2. It checks if the Solver has to find a candidate.
3. If so, it starts to search a matching Peering Candidate if available.
4. If some Peering Candidates are available, it selects one and book it.
5. If no Peering Candidates are available, it starts the discovery process by creating a `Discovery`.
6. If the `findCandidate` status is solved, it means that a Peering Candidate has been found. Otherwise, it means that the `Solver` has failed.
7. If in the `Solver` there is also a `ReserveAndBuy` phase, it starts the reservation process. Otherwise, it ends the process, the solver is already solved.
8. Firstly, it starts to get the `PeeringCandidate` from the `Solver` object. Then, it forge the Partition starting from the `Solver` selector. At this point, it creates a `Reservation` object.
9. If the `Reservation` is successfully fulfilled, it means that the `Solver` has reserved and purchased the resources. Otherwise, it means that the `Solver` has failed.
10. If in the `Solver` there is also a `EnstablishPeering` phase, it starts the peering process (to be implemented). Otherwise, it ends the process.

## Discovery Controller (`discovery_controller.go`)

The Discovery controller, tasked with reconciliation on the `Discovery` object, continuously monitors and manages its state to ensure alignment with the desired configuration. It follows the following steps:

1. When there is a new Discovery object, it firstly starts the discovery process by contacting the `Gateway` to discover flavours that fits the `Discovery` selector.
2. If no flavours are found, it means that the `Discovery` has failed. Otherwise, it refers to the first `PeeringCandidate` as the one that will be reserved (more complex logic should be implemented), while the other will be stored as not reserved.
3. It update the `Discovery` object with the `PeeringCandidates` found.
4. The `Discovery` is solved, so it ends the process.

## Reservation Controller (`reservation_controller.go`)

The Reservation controller, tasked with reconciliation on the `Reservation` object, continuously monitors and manages its state to ensure alignment with the desired configuration. It follows the following steps:

1. When there is a new `Reservation` object it checks if the `Reserve` flag is set. If so, it starts the **Reserve** process.
2. It retrieves the FlavourID from the `PeeringCandidate` of the `Reservation` object. With this information, it starts the reservation process through the `Gateway`.
3. If the reserve phase of the reservation is successful, it will create a `Transaction` object from the response received. Otherwise, the `Reservation` has failed.
4. If the `Reservation` has the `Purchase` flag set, it starts the **Purchase** process. Otherwise, it ends the process because the `Reservation` has already succeeded.
5. Using the `Transaction` object from the `Reservation`, it starts the purchase process.
6. If the purchase phase is successfully fulfilled, it will update the status of the `Reservation` object and it will store the received `Contract`. Otherwise, the `Reservation` has failed.

## Allocation Controller (`allocation_controller.go`)

The Allocation controller, tasked with reconciliation on the `Allocation` object, continuously monitors and manages its state to ensure alignment with the desired configuration.

(To be implemented)
