# gNMI Support for Multiple Client Roles and Master Arbitration

**Contributors:** Xinming Chen, Rob Shakir, Waqar Mohsin, Anees Shaikh, Tomek
Madejski, Alireza Ghaffarkhah

**Date:** April 20th, 2018

**Version:** 0.1.0

# 1 Master Arbitration

For high availability, a system may run multiple replicas of a gNMI client.
Among the replicas, only one client should be elected as master and do gNMI
operations that mutate any state on the target. However, in the event of a
network partition, there can be two or more replicas thinking themselves as
master. For read-only RPCs, this is OK. But if they both call the `Set` RPC, the
target may be incorrectly configured by the stale master. Therefore, "master
arbitration" is needed when multiple clients exist.

Master arbitration means when a gNMI client connects, it provides a 128-bit
integer `election_id`. The election ID MUST be guaranteed to be monotonically
increasing -- a new master's election ID MUST always be larger than its
predecessor. The process through which election IDs are allocated to client
replicas is out-of-scope of this specification.

The client carries the election ID in all of its `Set` requests. When the target
receives any `Set` request, it looks at the election ID carried with the
requests and stores the largest ID it sees. An election ID that is equal to or
greater than the currently stored one should be accepted as the master. On the
other hand, `Set` requests with no election ID or with a smaller election ID are
rejected and return an error.

# 2 Client Role

Sometimes, there is a need to partition the config tree among multiple clients
(where each client in turn could be replicated). This is accomplished by
assigning a *role* to each client. Master arbitration is performed on a per-role
basis. There can be one master for each role, i.e. multiple master clients are
allowed only if they belong to different roles.

The role is identified by its `id`, which is assigned offline in agreement
across the entire control plane. The target will use the role ID to arbitrate
such that each role has one and only one master controller instance.

## 2.1 Default Role

To simplify for use-cases where multiple roles are not needed, the client can
leave the role message unset in MasterArbitration. This implies default role.
All `Set` RPCs that uses a default role are arbitrated in the same group.


# 3 Implementation Detail

## 3.1 Proto Definition

A `MasterArbitration` message is embedded in the gNMI Extension message of the
`SetRequest` proto. It carries the election ID and the role. The message is
defined as following:

```
message MasterArbitration {
  Role role = 1;
  Uint128 election_id = 2;
}

message Uint128 {
  uint64 high = 1;
  uint64 low = 2;
}

message Role {
  string id = 1;
}
```

More fields can be added to the `Role` proto if needed, for example, to specify
what paths the role can read/write.

## 3.2 `SetRequest` Handling

In order to update the election ID as soon as a new master is elected, the
client is required to send an empty `Set` RPC (with the election ID only) as
soon as it becomes master. The client also carries the election ID in all
subsequent `Set` requests.

When the target receives a `Set` request, it compares the election ID with the
locally stored one within the same role:

-   If the election ID equals the currently stored one, proceeds with the Set
    operation.

-   If the election ID is larger than the currently stored one, updates the
    currently stored election ID with the larger one, then proceeds with the Set
    operation.

-   If the election ID is smaller than the currently stored one, returns
    `PERMISSION_DENIED` error. For debugging purposes, it is recommended that
    the target includes the currently stored highest election ID in the error
    message string.

-   If the arbitration extension exists but `election_id` is not set, returns
    `INVALID_ARGUMENT` error.

-   If there is no arbitration extension associated with the Set request, the
    Set request should be accepted and processed. Therefore, client that expects
    to ever be part of a replica-group MUST always set this extension, and the
    role MUST be populated.

-   If `role` message is not set, this request is considered as using the
    default role. It will be arbitrated against all the clients that have the
    default role.

