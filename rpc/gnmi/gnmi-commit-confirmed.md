# gNMI Commit Confirmed Extension

**Contributors:**

**Date:** October 6, 2023

**Version:** 0.1.0

## 1. Purpose

In certain deployments, client and server is separated by a complex network, hence we cannot assume

- The pushed configuration will not break connectivity to the network device.
- The network device has out-of-band access.

This extension provides a way for a controller/operator to apply updates in a `SetRequest` that will require a subsequent confirmation in a time window specified by the `timeout` value.

If no confirmation is received within the time window, the network device will revert the changes restoring the previous state. This workflow is particularly useful to ensure that erroneous configuration will not break connectivity to the network device.

## 2. Definition

A [`CommitConfirmed`](TBD) message is added to the list of well-known gNMI Extensions and is defined in [gnmi_ext.proto](TBD).

The message contains the following fields:

- `action` - The action to be taken by the network device upon receiving the `SetRequest` with the `CommitConfirmed` extension.
- `timeout` - time in seconds/nanoseconds to wait for a confirmation before reverting the changes.
- `commit_id` - commit identifier returned by the network device that is associated with the commit operation.

### 3. Set Request handling

#### 3.1 Initiating Commit Confirmed operation

To initiate a Commit Confirmed operation, the client MUST include the `CommitConfirmed` extension in the `SetRequest` message. The `CommitConfirmed` message MUST include the `COMMIT_CONFIRMED_ACTION_START` action to indicate that the network device MUST apply the configuration changes and start the commit rollback timeout waiting for a confirmation action.

The network device upon receiving the `SetRequest` with the `CommitConfirmed` extension with `COMMIT_CONFIRMED_ACTION_START` returns a `SetResponse` with `CommitConfirmed` extension message with the `commit_id` field set to the commit identifier associated with the commit. No other fields in the `CommitConfirmed` message are set.

If the `CommitConfirmed` message in the `SetRequest` does not include the `timeout` field, the network device uses its default commit confirmed timeout value and returns the `timeout` value in the `SetResponse` message in the `CommitConfirmed` extension.

If in the `SetRequest`, the `CommitConfirmed` message with `COMMIT_CONFIRMED_ACTION_START` action includes the `commit_id` field the network [resets the timeout](#34-resetting-the-timeout); else, if the targeted commit is not in progress device MUST return an error with the `FAILED_PRECONDITION` code.

#### 3.2 Confirming commit

To confirm a commit, the client MUST include the `CommitConfirmed` extension in the `SetRequest` message without any content in `update`/`delete`/`replace` fields. The `CommitConfirmed` message MUST include the `COMMIT_CONFIRMED_ACTION_CONFIRM` action and `commit_id` field set to the commit identifier returned by the network device in the `SetResponse` message with `COMMIT_CONFIRMED_ACTION_START` action.

As a result the network device should stop the commit rollback timeout which concludes the commit confirmed operation. Upon successful processing of the confirmation action the network device MUST NOT return `CommitConfirmed` extension in the associated `SetResponse` message.

If `COMMIT_CONFIRMED_ACTION_CONFIRM` is received for a commit that is not in progress, the network device MUST return an error with the `FAILED_PRECONDITION` code.

#### 3.3 Rejecting commit

To reject a commit for which the timeout timer has not yet expired, the client MUST include the `CommitConfirmed` extension in the `SetRequest` message without any content in `update`/`delete`/`replace` fields. The `CommitConfirmed` message MUST include the `COMMIT_CONFIRMED_ACTION_REJECT` action and `commit_id` field set to the commit identifier returned by the network device in the respecting `SetResponse` message.

As a result the network device should stop the commit rollback timeout and revert the commit identified by the provided `commit_id` which concludes the reject operation. Upon successful processing of the reject action the network device MUST NOT return `CommitConfirmed` extension in the associated `SetResponse` message.

If `COMMIT_CONFIRMED_ACTION_REJECT` is received for a commit that is not in progress, the network device MUST return an error with the `FAILED_PRECONDITION` code.

#### 3.4 Resetting the timeout

To reset the timeout for a commit for which the timeout timer has not yet expired, the client MUST include the `CommitConfirmed` extension in the `SetRequest` message without any content in `update`/`delete`/`replace` fields. The `CommitConfirmed` message MUST include the `COMMIT_CONFIRMED_ACTION_START` action, `timeout` value set to the new desired timeout, and `commit_id` field set to the commit identifier returned by the network device in the respecting `SetResponse` message.

The network device should reset the commit rollback timeout to the new value and return the `SetResponse` message without `CommitConfirmed` extension.

### 4. Commit ID

The commit identifier returned by the network device in the `SetResponse` message with `COMMIT_CONFIRMED_ACTION_START` action enables controlled session management and prevents erroneous actions to be taken by the network device due to race conditions.

For example, when commit_id is not used it is theoretically possible that the commit confirmed action will be processed by the network device after the timeout has expired, causing erroneous action to be taken. Or multiple actors will attempt to perform different actions on the same commit.
