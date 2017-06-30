## Schema path encoding conventions for gNMI

**Updated**: June 21, 2017

**Version**: 0.4.0

This document is a supplement to the [gNMI specification](https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-specification.md), and provides additional guidelines and examples for path encoding.

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

Path encoding in gNMI uses a structured path format. This format consists
of a set of elements which consist of a name, and an option map of keys.
Keys are represented as string values, regardless of their type within the
schema that describes the data. Some of the examples below assume YANG-
modeled data since it is a common use case for gNMI, but gNMI has
applicability to any data modeling where the data can be described in a
tree structure with hierarchical paths.

#### Constructing paths

gNMI paths are encoded as an ordered list (slice, array, etc.) of `gnmi.PathElem` messages (represented as a repeated field within the protocol buffer definition). Each `PathElem` consists of a name, encoded as a string. An element's name MUST be encoded as a UTF-8 string. Each `PathElem` may optionally specify a set of keys, specified as a `map<string,string>` (dictionary, or map). The key of the `key` map is the name of a key for the element, and the `value` represent the string-encoded value of the key. Both the key and value of the map MUST contain UTF-8 encoded strings.

*  the root path `/` is encoded as a zero length array (slice) of `PathElem` messages.  Example declarations in several languages:

    * Go: `path := []*PathElem{}`
    * Python: `path = []`
    * C++ : `vector<PathElem> path {};`

  Note this is not the same as a path consisting of a single empty
  string element.

*  a human-readable path can be formed by concatenating elements of the prefix
  and path using a `/` separator, and preceded by a leading `/` character.
  For example:

  ```
    prefix: <
    	elem: <
    		name: "a"
    	>
    >
    path: <
    	elem: <
    		name: "b"
    	>
    	elem: <
    		name: "c"
    	>
    >
  ```
  results in the path `/a/b/c` in human readable form.

*   subscription and data retrieval paths (Subscribe and Get RPCs) are
    recursive, i.e., select the specified element and all of its descendents. For example, the path:

  ```
  <
  	  elem: <
  		  name: "interfaces"
  	  >
  	    elem: <
  		  name: "interface"
  		  key: <
  			  key: "name"
  			  value: "Ethernet1/2/3"
  		  >
  	  >
  	  elem: <
  		  name: "state"
  	  >
  >
  ```

  represents all of the leaves
  in the state container, as well as leaves in descendent containers,
  e.g., including:

  ```
  /interfaces/interface[name=Ethernet1/2/3]/state/counters
  ```

* Each string within the `PathElem` message (i.e., the `name`, and the key and value of the `key` map) must contain a valid UTF-8 encoded string.

#### Paths referencing list elements

*   To reference a list element, both the `name` of the list and `key` map must be specified. i.e.,

  ```
  <
  		elem: <
  			name: "interfaces"
  		>
  		elem: <
  			name: "interface"
  			key: <
  				key: "name"
  				value: "Ethernet1/2/3"
  			>
  		>
  		elem: <
  			name: "state"
  		>
  		elem: <
  			name: "counters"
  		>
  	>
   ```

  selects the entry in the `interface` list with the `name` key specified to be `Ethernet1/2/3`.

* Where a list has multiple keys, each key is specified by an additional entry within the `key` map. For example:

  ```
  	<
  		elem: <
  			name: "network-instances"
  		>
  		elem: <
  			name: "network-instance"
  			key: <
  				key: "name"
  				value: "DEFAULT"
  			>
  		>
  		elem: <
  			name: "protocols"
  		>
  		elem: <
  			name: "protocol"
  			key: <
  				key: "identifier"
  				value: "ISIS"
  			>
  			key: <
  				key: "name"
  				value: "65497"
  			>
  		>
  	>
  	```

  * To match all entries within a particular list, the key(s) to the list may be omitted, for example to match the `oper-status` value of all interfaces on a device:

    ```
    <
    	elem: <
    		name: "interfaces"
    	>
    	elem: <
    		name: "interface"
    	>
    	elem: <
    		name: "state"
    	>
    	elem: <
    		name: "oper-status"
    	>
    >
    ```

    In this case, since the `interface` `PathElem` does not specify any keys, it should be interpreted to match all entries within the `interface` list. The same semantics can be specified by utilising an asterisk (`*`) for a particular `key` map entry's value, i.e.:

    ```
    <
    	elem: <
    		name: "interfaces"
    	>
    	elem: <
    		name: "interface"
    		key: <
    			key: "name"
    			value: "*"
    		>
    	>
    >
    ```

#### Wildcards in paths

*   Wildcards are allowed to indicate all elements at a given subtree in the
    schema -- these are used particularly for telemetry subscriptions or
    `Get` requests. A single-level wildcard is indicated by specifying the `name` of a `PathElem` to be an asterisk (`*`). A multi-level wildcard is indicated by specifying the `name` of a `PathElem` to be the string `...`.

*  Wildcards may be used in multiple levels of the path, e.g., select all elements named `state` three levels deep:

    ```
    <
    	elem: <
    		name: "interfaces"
    	>
    	elem: <
    		name: "*"
    	>
    	elem: <
    		name: "*"
    	>
    	elem: <
    		name: "*"
    	>
    	elem: <
    		name: "state"
    	>
    >
    ```

*   Per the specification above, wildcards may also appear as list key values.

  ```
  <
  	elem: <
  		name: "interfaces"
  	>
  	elem: <
  		name: "interface"
  		key: <
  			key: "name"
  			value: "*"
  		>
  	>
  	elem: <
  		name: "state"
  	>
  >
  ```

*   Wildcards should not appear in paths returned by a device in
    telemetry notifications.

*   A single path element, including keyed fields, maybe be replaced by
    `...` to select all matching descendents.  This is semantically equivalent
    to the empty element notation, `//`, in XPATH.

    ```
    <
    	elem: <
    		name: "interfaces"
    	>
    	elem: <
    		name: "interface"
    	>
    	elem: <
    		name: "..."
    	>
    >
 	```

    *   Select all `state` containers under interface 'eth0'
    ```
    <
    	elem: <
    		name: "interfaces"
    	>
    	elem: <
    		name: "interface"
    		key: <
    			key: "name"
    			value: "eth0"
    		>
    	>
    	elem: <
    		name: "..."
    	>
    	elem: <
    		name: "state"
    	>
    >
 	```

### Contributors
 * Paul Borman, Josh George, Kevin Grant, Chuan Han, Marcus Hines, Carl Lebsack, Anees Shaikh, Rob Shakir, Manish Verma (Google, Inc.)
 * Aaron Beitch (Arista)
 * Arun Satyanarayana (Cisco Systems)
 * Jason Sterne (Nokia)