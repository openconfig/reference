## Schema path encoding conventions for gNMI

**Updated**: February 24, 2016<br>
**Version**: 0.2.0

This document is a supplement to the [gNMI
specification](https://github.com/openconfig/reference/blob/master/rpc/gnmi
/gnmi-specification.md), and provides additional guidelines and examples for
path encoding.

### Path encoding

Data model path encoding and decoding is required in a gNMI-compliant
management stack.  For streaming telemetry, devices (targets) generate
notifications, or updates, with an associated path and value.  The
data collector must interpret these paths efficiently.  For device
configuration, the NMS client generates configuration data with an
associated path, and passes the data to the target, which must
correctly interpret the path against the appropriate data tree and
apply the configuration.

### Guidance for implementors

Path encoding in gNMI uses a variant of XPATH syntax with the
following rules.  Some of the examples below assume YANG-modeled data
since it is a common use case for gNMI, but gNMI has applicability to
any data modeling where the data can be described in a tree structure
with hierarchical paths.

#### Constructing paths

gNMI paths are encoded as a list (slice, array etc.) of path string elements.

*  the root path `/` is encoded as a zero length array (slice) of path
   elements.  Example declarations in several languages:

    * Go: `path := []string{}`
    * Python: `path = []`
    * C++ : `vector<string> path {};`

  Note this is not the same as a path consisting of a single empty
  string element.

*  a human-readable path can be formed by concatenating elements of the prefix
  and path using a `/` separator, and preceded by a leading `/` character.
  For example:
  ```
    prefix: []string{“a”}
    path: []string{“b”,"c"}
  ```
  results in the path `/a/b/c` in human readable form.

*   subscription and data retrieval paths (Subscribe and Get RPCs) are
    recursive, i.e., select the specified element and all of its descendents

  `/interfaces/interface[name=Ethernet1/2/3]/state` -- all of the leaves
  in the state container, as well as leaves in descendent containers,
  e.g., including:

  ```
  /interfaces/interface[name=Ethernet1/2/3]/state/counters
  ```

*   newlines and carriage returns are encoded as \n, \r, respectively.
    A tab is optionally encoded as \t.  Non-ascii values should be encoded as
    \x12 (value 18).  Unicode characters are encoded as \u1234 or \U12345678.

#### Paths referencing list elements

*   list names and keys appear in the same path element

  `/interfaces/interface[name=Ethernet1/2/3]/state/counters` -- all of
  the counter leaves on the named interface

  `/interfaces/interface[name=*]/state/oper-status` -- operational status
  for all interfaces

  `/bgp/neighbors` -- contents of neighbor list, i.e. all neighbors

  `/bgp/neighbors/neighbor[neighbor-address=172.24.3.10]` -- all data for
  the corresponding BGP neighbor (i.e. the entire subtree)

*   in paths with list keys, only closing square brackets and \ characters must
    be escaped if they are part of a key value.  The = character may not appear
    in key names in YANG, so `[name=k1=v1]` is not ambiguous, i.e., `name` is
    assigned the value of `k1=v1`.  Further, the sequence `[name=[foo]` is not
    ambiguous,  name is assigned the value of `[foo`.  On the other hand setting
    name to the value `[\]` would be encoded as `[name=[\\\]]`.
    Escape can optionally be applied to the opening square brackets.

*   values in list key selectors are not enclosed in quotes (unless
    the quotes are actually part of the value).  The value extends
    from the first `'='` to the closing (unescaped) `']'`.  When the
    type of the value (integer, string, etc.) must be known, e.g. for
    validation, it can be recovered from the corresponding schema.

*   list keys and values must appear in the order defined in the corresponding
    schema.  All list keys must appear.
  ```
    /network-instances/network-instance/tables/table[protocol=BGP][address-family=IPV4]
  ```

#### Wildcards in paths

*   wildcards are allowed to indicate all elements at a given subtree in the
    schema -- these are used particularly for telemetry subscriptions or
    `Get` requests.

  *  wildcards may be used in multiple levels of the path, e.g., select all state
     containers three levels deep
    ```
    /interfaces/*/*/*/state
    ```

*   wildcards may also appear as list key values.
  ```
    /interfaces/interface[name=*]/state
  ```
    Note that all list keys must appear, even if some are wildcarded:
  ```
  /network-instances/network-instance/tables/table[protocol=BGP][address-family=*]
  ```
  This could be relaxed in future revisions of this specification by allowing
  omission of wildcarded list keys.

*   wildcards should not appear in paths returned by a device in
    telemetry notifications.

*   a single path element, including keyed fields, maybe be replaced by
    `...` to select all matching descendents.  This is semantically equivalent
    to the empty element notation, `//`, in XPATH.
    ```
      /interfaces/interface/.../state
    ```
    *   Select all `state` containers under interface 'eth0'
    ```
      /interfaces/interface[name=eth0]/.../state
    ```
    *   Select all state attributes under the the interface `config` or `state`
        containers for all interfaces
    ```
      /interfaces/interface[name=*]/.../state
      /interfaces/interface[name=*]/.../config
    ```

**Contributors**: Paul Borman, Josh George, Kevin Grant, Chuan Han, Marcus Hines, Carl Lebsack, Anees Shaikh, Rob Shakir, Manish Verma
