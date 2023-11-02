# gNMI Commit Confirmed Extension

**Contributors:** Gautham V Kidiyoor, Roman Dodin, Rob Shakir, Vinit Kanvinde, Priyadeep Bangalore Lokesh

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
// Commit confirmed extension allows automated revert of the configuration after
// certain duration if an explicit confirmation is not issued. It allows explicit
// cancellation of the commit during the rollback window. There cannot be more
// than one commit active at a given time.
// The document about gNMI commit confirmed can be found at
// https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-commit-confirmed.md
message Commit {
  // ID is provided by the client during the commit request. During confirm and cancel
  // actions the provided ID should match the ID provided during commit.
  // If ID is not passed in any actions server shall return error.
  // Required.
  string id = 1;
  oneof action {
    // commit action creates a new commit. If a commit is on-going, server returns error.
    CommitRequest commit = 2;
    // confirm action will confirm an on-going commit, the ID provided during confirm
    // should match the on-going commit ID.
    CommitConfirm confirm = 3;
    // cancel action will cancel an on-going commit, the ID provided during cancel
    // should match the on-going commit ID.
    CommitCancel cancel = 4;
  }
}

// CommitRequest is used to create a new confirmed commit. It hold additional
// parameter requried for commit action.
message CommitRequest {
  // Maximum duration to wait for a confirmaton before reverting the commit.
  google.protobuf.Duration rollback_duration = 1;
}

// CommitConfirm is used to confirm an on-going commit. It hold additional
// parameter requried for confirm action.
message CommitConfirm {}

// CommitCancel is used to cancel an on-going commit. It hold additional
// parameter requried for cancel action.
message CommitCancel {}
```

## 3.2 SetRequest handling                                                        

### 3.2.1 Commit
A commit can be initiated by providing `CommitRequest` as action in the extension. A `id` must to be
provided by the client. The server shall associate the commit with the provided `id`.
During confirm or cancel action the provided `id` must match the `id` of the on-going commit.
`rollback_duration` can be used to override the default rollback duration which is 10min.
If a confirmation call is not received before the rollback duration then the configuration is reverted.

If a commit is issued whilst an existing rollback counter is running then the server returns with
FAILED_PRECONDITION error.

If a SetRequest call is made without extension whilst an existing rollback counter is running then a
FAILED PRECONDITION error is returned.

### 3.2.2 Confirm

Confirmation can be issued by providing `ConfirmRequest` as action in the extension. The value of `id`
should be equivalent to the `id` of the on-going commit.

If the server is not waiting for confirmation or if the value doesn’t match the on-going commit then
FAILED_PRECONDITION or INVALID_ARGUMENT error is returned respectively.

### 3.2.3 Cancel
Cancellation can be issued by providing `CancelRequest` as action in the extension. The value of `id`
should be equivalent to the id of the on-going commit.

If the server is not waiting for cancellation or if the value doesn’t match the on-going commit
then FAILED_PRECONDITION or INVALID_ARGUMENT error is returned respectively.
