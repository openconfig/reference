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
// Commit confirmed extension allows automated revert of a configuration after
// certain duration if an explicit confirmation is not issued. This duration can
// be used to verify the configuration.
// It also allows explicit cancellation of the commit during the time window of rollback
// duration if client determine that the configuration needs to be reverted.
// The document about gNMI commit confirmed can be found at
// https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-commit-confirmed.md
message Commit {
  oneof action {
    CommitRequest commit = 1;
    CommitConfirm confirm = 2;
    CommitCancel cancel = 3;
  }
}

// Create a new commit request.
message CommitRequest {
  // Commit id for the request.
  string id = 1;
  // Maximum duration to wait for a confirmaton before reverting the commit.
  google.protobuf.Duration rollback_duration = 2;
}

// Confirm an on-going commit.
message CommitConfirm {
  // Commit id that requires confirmation.
  string id = 1;
}

// Cancel an on-going commit.
message CommitCancel {
  // Commit id that requires cancellation.
  string id = 1;
}
```

## 3.2 SetRequest handling                                                        

### 3.2.1 Commit
A commit can be initiated by setting `CommitRequest` as action in the extension. A commit `id` needs 
to be provided which will be used during confirmation or cancellation. `rollback_duration` can be used
to override the default rollback duration which is 10min. If a confirmation call is not received before
the rollback duration then the configuration is reverted.

If a commit is issued whilst an existing rollback counter is running then the server returns with
FAILED_PRECONDITION error.

If a SetRequest call is made without extension whilst an existing rollback counter is running then a
FAILED PRECONDITION error is returned.

### 3.2.2 Confirm

Confirmation can be issued by setting `ConfirmRequest` as action in the extension. The value of `id`
should be equivalent to the commit id of the on-going commit which needs confirmation.

If the server is not waiting for a confirmation or if the value doesn’t match the on-going commit then
FAILED_PRECONDITION or INVALID_ARGUMENT error is returned respectively.

### 3.2.3 Cancel
Cancellation can be issued by setting `CancelRequest` as action in the extension. The value of `id`
should be equivalent to the commit id of the on-going commit which needs cancellation.

If the server is not waiting for a cancellation or if the value doesn’t match the on-going commit
then FAILED_PRECONDITION or INVALID_ARGUMENT error is returned respectively.
