# gNMI History Extension

**Contributors:** Justin Costa-Roberts, Aaron Beitch, Mike Fink

**Date:** August 4, 2020

**Version:** 0.1.0

# 1. Purpose

In certain applications it is necessary not only to retrieve current data from a
data tree, but also to allow querying historical data. The gNMI History
extension enables clients both to retrieve data at a specific time in the past,
and to request all updates applied to a data tree between two specified times.

# 2. Definition

A `History` message is embedded in an Extension message in the
`SubscribeRequest` proto.

## 2.1 Proto

```
message TimeRange {
  int64 start = 1;
  int64 end = 2;
}

message History {
  oneof request {
    int64 snapshot_time = 1;
    TimeRange range = 2;
  }
}
```

## 2.2 SubscribeRequest handling

The History extension may be applied only to `ONCE` and `STREAM` subscriptions.
Requests for historical data can be either of two types: a snapshot at a
specified time, or a set of updates between two times. The `History` message's
request field is a `oneof` field whose type indicates which of these request
types applies.

In all cases, times are represented as int64s interpreted as nanoseconds since
the Unix Epoch.

### 2.2.1 Snapshot requests

If the `snapshot_time` field is set, the request is for a snapshot. The
subscription mode for a snapshot request must be `ONCE`; otherwise the server
must return an `INVALID_ARGUMENT` error. The server must transmit the data
indicated in the `SubscribeRequest` as it existed on the server at the specified
time. If the specified time indicates a future time at the point the server
receives the request, the server may either return an `UNIMPLEMENTED` error or
choose to wait until the specified time has been reached.

### 2.2.2 Range requests

If the `range` field is set, the request is for the set of updates that occurred
during the half-open interval between the specified start and end times. Such a
request may return multiple updates per leaf, so the subscription mode for a
range request must be `STREAM`. If the mode of a range request is not `STREAM`,
the server should return an `INVALID_ARGUMENT` error. For the paths specified in
the `SubscriptionList`, the server must transmit all updates it's aware of that
occurred at the start time or later and before the end time specified in the
`TimeRange`, and then close the RPC. As usual, the server should also transmit
the specified data tree as it existed at the subscription start time unless the
`updates_only` field on the `SubscriptionList` message is set.

If `TimeRange`'s start time is later than its end time, the server must return
an `INVALID_ARGUMENT` error. If a start or end time has not yet elapsed when the
server receives the request, the server may either return an `UNIMPLEMENTED`
error or honor the request by waiting until the specified time. Clients may
therefore request open-ended `STREAM` subscriptions beginning in the past by
setting the end time to, for example, the largest int64 value.

### 2.2.3 Considerations for requests not immediately serviceable

Clients should not expect servers to service requests with start or end times in
the future. Servers may honor such requests in order to account for the
possibility of moderate differences in local time between the client and server,
but clients should expect that times far into the future will not be serviced,
and that significant clock desynchronization may cause the server to return an
error for requests with specified times near the present. Finally, clients
should expect that in order not to exhaust resources, servers may limit the
number of open requests that cannot immediately be serviced.
