# gNMI path encoding conventions

## Schema path encoding conventions for gNMI
*Updated*: October 13, 2016
*Version*: 0.2.1


This document is a supplement to the gNMI specification, and provides additional guidelines and examples for path encoding.

### Path encoding

Data model path encoding and decoding is required gNMI-compliant management stack.  For streaming telemetry, devices (targets) generate notifications, or updates, with an associated path and value.  The data collector must interpret these paths efficiently.  For device configuration, the NMS client must generate configuration data with an associated path, and pass the data to the target, which must correctly interpret the path against the appropriate data tree and apply the configuration.

### Guidance for implementors

Path encoding in gNMI uses a variant of XPATH syntax with the following rules. where list name and keys appear in the same path element.  Some of the examples below assume YANG-modeled data since it is a common use case for gNMI, but gNMI has applicability to any data modeling where the data can be described in a tree structure.



*   subscription paths select the specified element and all of its descendents (i.e., is recursive)

  `/interfaces/interface[name=Ethernet1/2/3]/state` -- all of the leaves in the state container, as well as leaves in descendent containers, e.g.:

  ```
  /interfaces/interface[name=Ethernet1/2/3]/state/counters
  ```

*   list names and keys appear in the same path element

  `/interfaces/interface[name=Ethernet1/2/3]/state/counters` -- all of the counter leaves on the named interface
  `/interfaces/interface[name=*]/state/oper-status` -- operational status for all interfaces
  `/bgp/neighbors` -- contents of neighbor list
  `/bgp/neighbors/neighbor[neighbor-address=172.24.3.10]` -- all data for the corresponding BGP neighbor (i.e. the entire subtree)

*   in list paths, only closing square brackets and \ must be escaped.  The = character may not appear in key names in YANG, so `[name=k1=v1]` is not ambiguous, `name` is assigned the value of `k1=v1`.  Further, the sequence `[name=[foo]` is not ambiguous, name is assigned the value of `[foo`.  On the other hand setting name to the value` [\]` would be encoded as `[name=[\\\]]`.  Escape can optionally be applied to the opening square brackets.

*   values in list key selectors are not enclosed in quotes (unless the quotes are actually part of the value).  The value extends from the first `'='` to the closing (unescaped) `']'`.  When the type of the value must be known, e.g. for validation, it can be recovered from the corresponding schema.

*   newlines and carriage returns are encoded as \n, \r, respectively.  A tab is optionally encoded as \t.  Non-ascii values should be encoded as \x12 (value 18).  Unicode characters are encoded as \u1234 or \U12345678.

*   wildcards are allowed to indicate all elements at a given subtree in the schema -- these are used particularly for telemetry subscriptions or `Get` requests; wildcards may also appear as list key values (e.g., `.../interfaces/interface[name=*]/state`).  Wildcards should not appear in paths returned by a device in telemetry notifications.

*   list keys and values must appear in the order defined in the corresponding schema
  *   all list keys must appear, even if some are wildcarded (i.e., don't cares), e.g.:
  ```
  /network-instances/network-instance/tables/
		table[protocol=BGP][address-family=*]
  ```

This could be relaxed in future revisions of this specification by allowing omission of wildcarded list keys.

*   any element, including keyed fields, maybe be replaced by the `//` wildcard
    *   Select all state containers under all interfaces
        *   `.../interfaces/interface//state`
    *   Select all state containers under interface 'eth0'
        *   `.../interfaces/interface[name=eth0]//state`
    *   Select all state containers three levels deep
        *   `.../*/*/*/state`
    *   Select all state attributes under the the interface config or state containers for all interfaces
        *   `.../interfaces/interface[name=*]//state .../interfaces/interface[name=*]//config`
