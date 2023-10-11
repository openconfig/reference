# gNMI Commit Confirmed Extension

**Contributors:** Gautham V Kidiyoor, Rob Shakir, Vinit Kanvinde, Priyadeep Bangalore Lokesh

**Date:** September 20, 2023

**Version:** 0.1.0

# 1. Purpose

In certain deployments, client and server is seperated by a complex network,
hence we cannot assume
- The pushed configuration will not break connectivity to the network device.
- The network device have out-of-band access.

This feature provides a way to auto rollback the applied configuration after a
centain duration if a bad configuration was pushed.

# 2. Summary
The proposed proto has a subset of confirmed commit functionality as defined in
NETCONF protocol([RFC6241](https://datatracker.ietf.org/doc/html/rfc6241#section-8.4)). The proposal has a healthy disregard to few functionality
defined in the RFC with the intention that most of the gRPC API clients are going to
be automated systems and the proto should facilitate a simpler implementation of
client workflows and server implementation.

The server can be viewed as a singleton resource, at any given point in time there
can be only one commit active. This commit can be either confirmed or canceled
before a new commit can begin. Client is expected to provide full configuration
during the commit request, the commit cannot be amended once issued.

# 3. Definition

A `Commit` message is embedded the Extension message of the SetRequest proto.

## 3.1 Proto

```
message Commit {
  google.protobuf.Duration rollback_duration = 1;
  oneof id { 
   string commit_id = 2;
   string confirm_id = 3;
   string cancel_id = 4;
  }
}
```

## 3.2 SetRequest handling                                                        

### 3.2.1 Commit
A commit can be initiated by setting the `rollback_duration` field and `commit_id` in the extension.
If `commit_id` is passed without `rollback_duration` a default duration of 10min is chosen.

If the server is already waiting for a confirmation, the server returns with FAILED_PRECONDITION error.

If `rollback_duration` is passed without `commit_id`, an INVALID_ARGUMENT error is returned.

If a SetRequest call is made without extension whilst the existing rollback counter is running then a
FAILED PRECONDITION error is returned.

### 3.2.2 Confirm

Confirmation can be issued by setting the `confirm_id` to a value equivalent to the commit id of the
on-going commit which needs confirmation.

If the server is not waiting for a confirmation or if the value doesn’t match the on-going commit then
FAILED_PRECONDITION or INVALID_ARGUMENT error is returned respectively.

### 3.2.3 Cancel
Cancellation can be issued by setting the `cancel_id` to a value equivalent to the commit id of the
on-going commit.

If the server is not waiting for a confirmation or if the value doesn’t match the on-going commit
then FAILED_PRECONDITION or INVALID_ARGUMENT error is returned respectively.
