# gNMI Config Subscription Extension

**Contributors:** Roman Dodin, Matthew MacDonald

**Date:** July 29th, 2024

**Version:** 0.1.0

## 1. Purpose

Performing configuration management and handling configuration drift is one of
the main features of a higher-level management system or orchestrator.
Configuration drift occurs when a device's configuration changes independently
of the expected configuration defined by the management system. Such changes
may arise from manual interventions, automated processes outside the management
system, or unexpected behavior. When drift is detected, the management system
can either revert the configuration back to the expected configuration or
incorporate the change by updating its records to include the new configuration
as the expected configuration. The configuration management tasks focus on the
effective retrieval and push of configuration values.

Thus, having a synchronized configuration view between the management
system and the network elements is key to enabling robust and near real-time
configuration management.
To enable this synchronization of configuration data, the gNMI Subscribe RPC
can be used. The bidirectional streaming nature of this RPC enables fast
and reliable sync between the management system and the devices it manages.

Unfortunately, gNMI Subscribe RPC does not have an embedded mechanism to
stream updates for the configuration values only as opposed to the Get RPC,
which makes this RPC rather ineffective on YANG schemas that do not employ
a separation of config and state elements by using distinct containers.

This proposal introduces the Config Subscription extension that allows clients
to indicate to servers that they are interested in configuration values only.

gNMI Config Subscription Extension proto specification is defined in
[gnmi_ext.proto](
https://github.com/openconfig/gnmi/blob/master/proto/gnmi_ext/gnmi_ext.proto
).

## 2. Definition

A `ConfigSubscription` message is embedded as an extension message in the
`SubscribeRequest` or `SubscribeResponse` proto messages.
If the extension is embedded in a `SubscribeRequest`, the action field
must be a `ConfigSubscriptionStart`.
The presence of such an extension indicates to the target that the client
wants to start a ConfigSubscription. The target must return notifications
pertaining to data leaves that the target considers to be writable.
If the subscription type is `ON_CHANGE`, the target must separate the
notifications triggered by different commits using a
`ConfigSubscriptionSyncDone` in a `SubscribeResponse` message.
On the other hand, if the extension is embedded in a `SubscribeResponse`, the 
action field must be a `ConfigSubscriptionSyncDone`. This action is used by a
target to indicate a commit boundary to the client.

## 2.1 Proto

The extension contains a message called `ConfigSubscription` that carries
one of 2 types of actions. `ConfigSubscriptionStart` or
`ConfigSubscriptionSyncDone`

```proto
// ConfigSubscription extension allows clients to subscribe to configuration
// schema nodes only.
message ConfigSubscription {
  oneof action {
    // ConfigSubscriptionStart is sent by the client in the SubscribeRequest
    ConfigSubscriptionStart start = 1;
    // ConfigSubscriptionSyncDone is sent by the server in the SubscribeResponse
    ConfigSubscriptionSyncDone sync_done = 2;
  }
}

// ConfigSubscriptionStart is used to indicate to a target that for a given set
// of paths in the SubscribeRequest, the client wishes to receive updates
// for the configuration schema nodes only.
message ConfigSubscriptionStart {}

// ConfigSubscriptionSyncDone is sent by the server in the SubscribeResponse
// after all the updates for the configuration schema nodes have been sent.
message ConfigSubscriptionSyncDone {
  // ID of a commit confirm operation as assigned by the client
  // see Commit Confirm extension for more details.
  string commit_confirm_id = 1;
  // ID of a commit as might be assigned by the server
  // when registering a commit operation.
  string server_commit_id = 2;
  // If true indicates that the server is done processing the updates related to the
  // commit_confirm_id and/or server_commit_id.
  bool done = 3;
}
```

## 2.2 Actions

### 2.2.1 ConfigSubscriptionStart

A `ConfigSubscriptionStart` message is used by a gNMI client in a
`SubscribeRequest` to indicate that it wants to start a ConfigSubscription.
The target must respond exclusively with configuration data relevant to the
created subscription.

The base behavior of the `Subscribe` RPC remains unchanged: the target must
respond with an initial set of updates, followed by a `SubscribeResponse`
with the `sync_response` field set to true. However, if the updates_only
field in the SubscribeRequest is set to true, the target should omit the
initial updates and instead send only a SubscribeResponse with `sync_response`
set to true, in accordance with the gNMI specification.

### 2.2.2 ConfigSubscriptionSyncDone

The `ConfigSubscriptionSyncDone` message is included in a `SubscribeResponse`
by a gNMI target to signify a commit boundary to the client.
A commit boundary represents the completion of a specific set of changes
associated with a commit. It confirms that all changes in the set have been
successfully committed, with no errors reported by the interface used for the
commit, such as gNMI, NETCONF, CLI, etc.
[See Appendix - concurrent commits handling](#appendix-concurrent-commits-handling)

The `ConfigSubscriptionSyncDone` message includes three fields:

* `commit_confirm_id`: A commit confirm ID assigned by the client which
initiated the commit.
The commit can be initiated via gNMI (using the CommitConfirmed Extension),
NETCONF, or any other management interface. Applicable only if the commit
confirmed option is used.
* `server_commit_id`: An optional internal ID assigned by the target.
* `done`: If true, indicates that the server is done processing the updates
related to the commit_confirm_id and/or server_commit_id.

In the case a commit happens before the `sync_response: true` the server
cannot send a `ConfigSubscriptionSyncDone` until the `sync_response: true` has
been sent. The server may send the committed changes updates before the
`sync_response: true`.

## 3. Configuration changes scenarios

### 3.1 Configuration changes without Commit Confirmed

1) The client subscribes to path P1 with the `ConfigSubscription` extension
present with the action `ConfigSubscriptionStart`.
2) The server processes the subscription request as usual but will only send
updates for the configuration schema nodes under the path P1.
3) The client sends a Set RPC with the configuration changes to the path P1
**without** the `CommitConfirm` extension.
4) The server processes the Set RPC as usual and sends the updates for the
configuration schema nodes under the path P1.
5) After all the configuration updates are sent, the server sends the
`ConfigSubscriptionSyncDone` message to the client in a SubscribeResponse
message.
This message does not contain a `commit_confirmed_id` and may contain a
`server_commit_id`

### 3.2 Configuration changes with Commit Confirmed

1) The client subscribes to the path P1 with the `ConfigSubscription`
extension present with the action `ConfigSubscriptionStart`.
2) The server processes the subscription request as usual but will only send
updates for the configuration schema nodes under the path P1.
3) The client sends a Set RPC with the configuration changes to the path
P1 and **with** the `CommitConfirm` extension present.
4) The server processes the Set RPC as usual and sends the updates for
the configuration schema nodes under the path P1.
5) As all the configuration updates are sent, the server sends the
`ConfigSubscriptionSyncDone` message to the client in a SubscribeResponse
message.
This message must contain the the value of the `commit_confirmed_id`
received in the Set RPC in step 4 and may contain a `server_commit_id`.
6) When the client sends the commit confirm message, the server confirms
the commit and does not send any extra SubscribeResponse messages with the
`ConfigSubscriptionSyncDone` message.

### 3.3 Configuration changes with Commit Confirmed and rollback/cancellation

1) The client subscribes to path P1 with the `ConfigSubscription` extension
present with the action `ConfigSubscriptionStart`.
2) The server processes the subscription request as usual but will only send
updates for the configuration schema nodes under the path P1.
3) The client sends a Set RPC with the configuration changes to the path P1
and **with** the `CommitConfirm` extension present.
4) The server processes the Set RPC as usual and sends the updates for the
configuration schema nodes under the path P1.
5) After all the configuration updates are sent, the server sends the
`ConfigSubscriptionSyncDone` message to the client in a SubscribeResponse
message.
This message must contain the the value of the `commit_confirmed_id` received
in the Set RPC in step 4.
6) When the commit confirmed rollback timer expires or a commit cancel message
is received, the server:
  i. rolls back the configuration changes as per the Commit Confirm extension
  specification.
  ii. sends the new configuration updates for the path P1 as the configuration
  has changed/reverted.
  iii. sends the ConfigSubscriptionSyncDone message to the client in a
  `SubscribeResponse` message.
  This message must contain the the value of the `commit_confirmed_id`
  received in the Set RPC in step 4 and may contain a `server_commit_id`.

## Appendix: Concurrent commits handling

### Overlapping Update Streams

In scenarios where updates from multiple configuration changes overlap during 
streaming, it is important to note that the `ConfigSubscriptionSyncDone` 
message does not guarantee that the state reflected by received updates 
corresponds exclusively to the commit referenced by the 
`ConfigSubscriptionSyncDone` message. For example:

1. Timeline:

t = 0: Configuration change 1 is committed.
t = 1: Streaming updates for configuration change 1 begins.
t = Y (Y < N): Configuration change 2 is committed.
t = Y + 1: Streaming updates for configuration change 2 begins.
t = N: Streaming updates for configuration change 1 ends.
t = Y + N: Streaming updates for configuration change 2 ends.

2. Implications:

* Updates for paths impacted by configuration change 2 may be sent 
to the client before all updates for configuration change 1 are 
fully streamed.
* When the client receives a `ConfigSubscriptionSyncDone` message 
for configuration change 1, it might already have received updates
reflecting changes introduced by configuration change 2.
* Coalescing updates (e.g., combining overlapping changes to a path like `/foo/bar`)
If both configuration changes 1 and 2 modify `/foo/bar`, and the update for 
`/foo/bar` is delayed, the server may send only the final state of `/foo/bar` 
after configuration change 2. This approach optimizes performance but results
in the client never receiving the updates for change 1.

### Guarantees

The `ConfigSubscriptionSyncDone` message indicates that the server has completed 
processing all changes associated with the referenced commit. 
It does not guarantee that the client has received updates exclusively for that commit.

If a client requires strict state guarantees, it must implement its own 
mechanisms to lock configuration changes between commits. For example:

* Pausing further commits until the client confirms that all updates for the 
current commit are processed. 
* Implementing a client-side commit locking mechanism 
to avoid overlapping streams.
