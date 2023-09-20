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

# 2. Definition

A `Commit` message is embedded the Extension message of the SetRequest proto.

## 2.1 Proto

```
message Commit {
  google.protobuf.Duration rollback_duration = 1;
  bool confirm = 2;
}
```

## 2.2 SetRequest handling

### 2.2.1 Commit
If the *rollback_duration* field is set in SetRequest RPC, the server shall commit the configuration and wait until the
specified duration to initiate a rollback of the configuration unless there is another SetRequest RPC with *confirm* set
to true. If the second call is not received with confirm set to true the server shall rollback the committed
configuration.

### 2.2.2 Confirm

During the *confirm* call the client should send the same configuration in the SetRequest spec along with
the *confirm* field set to true.

If the confirm call is issued with a different configuration, a FAILED_PRECONDITION error is returned.

The confirm call can only be issued if the server is waiting for confirmation. If a SetRequest RPC is received with the
*confirm* field set to true but the server is not waiting for a confirmation then an FAILED_PRECONDITION error is
returned.

### 2.2.3 Multiple SetRequest

If a subsequent SetRequest RPC is received with the *rollback_duration* field set, whilst an existing rollback counter
is running, the server shall return a FAILED_PRECONDITION error.

If a SetRequest call is made without extension whilst the existing rollback counter is running then a
FAILED PRECONDITION error is returned.

