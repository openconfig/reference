# Mixing Schemas in gNMI

## Problem Statement

Today's network devices support multiple schema representations of their
underlying configuration and telemetry databases. Particularly, devices
typically support:

* A "native" schema. This representation generally 1:1 maps to the underlying
schema of the database storing configuration and telemetry variables on the
device.
* A vendor-neutral schema. This representation transforms the underlying schema
into a vendor-neutral model, namely OpenConfig.
* A CLI schema. This represents the text CLI format that has
been used to configure the device by humans.

Network operators transitioning to using modeled data for network configuration
or telemetry  must often make a partial transition.  This is typically due to
implementation support on the target device being provided in phases -- with a
mix of native models, vendor-neutral models, or CLI in a given version of the
software.

Additionally, some implementations of gNMI may allow multiple underlying
applications - each of which has its own schema - to be addressed. For example,
allowing configuration of particular NPUs, or other daemons on the device.

To allow a single gNMI interface to be used to address these different sources
of configuration and telemetry data, there is a requirement to allow
disambiguation of these different sets of content. Particularly, the
requirement is to allow disparate (and potentially overlapping) schema trees to
be addressed.

## Implementation

### Origin Specification in `Path`

In order to address the mixed-schema use case in gNMI, an `origin` field is
introduced to the `Path` message. It is encoded as a string. The path specified
within the message is uniquely identified by the tuple of `<origin, path>`.

The `origin` field is valid in any context of a `Path` message. Typically it is
used:

* In a `SetRequest` to indicate a particular schema is being used to modify
  the target configuration.
* In a `GetRequest` to retrieve the contents of a particular schema, or in
  a `GetResponse` to indicate that the payload contains data from a
  particular `<origin, path>` schema.
* In a `SubscribeRequest` to subscribe to paths within a particular schema,
  or `SubscribeResponse` to indicate an update corresponds to a particular
  `<origin, path>` tuple.

If more than one `origin` is to be used within any message, a path in the
`prefix` MUST NOT be specified, since a prefix applies to all paths within the
message. In the case that a `prefix` is specified, it MUST specify any required
`origin`. A single request MUST NOT specify `origin` in both `prefix` and `path`
fields in any RPC payload messages.

### Special Values of `origin`

Origin values are agreed out of band to the gNMI protocol. Currently, special
values are registered within this section of this document. Should additional
origins be defined, a registry will be defined.

Where the `origin` field is unspecified, its value should default to
`openconfig`. It is RECOMMENDED that the origin is explictly set.

Where the `origin` field is specified as `cli`, the path should be ignored and
the value specified within the `Update` considered as CLI input.

### Definition of `origin` for YANG-modelled Data

The `openconfig-extensions:origin` field MAY be utilised to determine the
origin within which a particular module is instantiated. The specification of
this extension is within
[openconfig-extensions.yang](https://github.com/openconfig/public/blob/master/release/models/openconfig-extensions.yang).

It should be noted that `origin` is distinct from `namespace`. Whilst a YANG
namespace is defined at any depth within the schema tree, an `origin` is
only used to disambiguate entire schema trees. That is to say, any element
that is not at the root inherits its `origin` from its root entity, regardless
of the YANG schema modules that make up that root.

### Partial Specifications of Origin in Set

If a `Set` RPC specifies `delete`, `update`, or `replace` fields which include
an `origin` within their `Path` messages, the corresponding change MUST be
constrained to the specified origin. Particularly:

* `replace` operations MUST only replace the contents of the specified `origin`
  at the specified path. Origins that are not specified within the `SetRequest`
  MUST NOT have their contents replaced. In order for a `replace` operation to
  replace any contents of an `origin` it must be explicitly specified in the
  `SetRequest`.
* `delete` operations MUST delete only the contents at the specified path within
  the specified `origin`. To delete contents from multiple origins, a client
  MUST specify multiple paths within the `delete` of the `SetRequest`.

### Special considerations for the use of Origin CLI

The following rules apply to a `SetRequest` containing `origin` CLI.  TODO:
Special considerations for mixing native schema are currently undefined.

* If a `SetRequest` contains `origin` `cli` and `origin` `openconfig`, the CLI
  MUST be given precedence over the OpenConfig origin.  This precedence is
  independent of the order in which origin CLI appears.  This overrides the
  gNMI specification requirement of ordered path based precedence within a
  `SetRequest`.  The goal of this precedence is to create deterministic
  behavior for resolving overlapping configuration.  For setRequests without
  `origin` CLI, order based precedence still applies.
* The `update` operation MUST be supported for setRequests containing `origin`
  `cli` and `origin` `openconfig`.  The `update` operation is compatible with
  the requirements for `origin` `cli` to take precedence.
* The `replace` operation for a setRequest containing only `origin` `cli` MUST
  be treated as a full device configuration replacement.
* The `replace` operation for a setRequest containing both CLI and OpenConfig
  SHOULD be supported.  The content of the two origins
  SHOULD be merged, with precedence to CLI for any overlapping
  configuration, and then the replace operation performed.  Since the CLI is
  interpreted as a full configuration replacement, the result of this
  setRequest should be a full configuration replacement with the merged
  content of the CLI and OpenConfig paths.
* A `replace` operation for `origin` `openconfig`, followed by a `update`
  operation for `origin` `cli` in a single setRequest SHOULD be supported.
  The goal is to allow configuration replacement to occur at a sub-tree level
  using both OpenConfig and the CLI configuration data.

### Transactionality of Sets with multiple Origins

Where a `SetRequest` specifies more than one `origin` - i.e., two or more
operations whose path include more than one origin - manipulations to all
affected trees MUST be considered as a single transaction. That is to say, only
if all transactions succeed should the `SetResponse` indicate success. If any
of the transactions fail, the contents of all origins MUST be rolled back, and
an error status returned upon responding to the `Set` RPC.

## Example

### Example: OpenConfig and CLI Data

If a client wishes to replace OpenConfig and update CLI-modelled data
concurrently, it can send the following `SetRequest`:

```
replace: <
  path: <
    origin: "openconfig"
  >
  val: <
    json_val: `
      {
        "interfaces": {
          "interface": [
            {
              "name": "eth0",
              "config": {
                "name": "eth0",
                "admin-status": "UP"
              }
            }
          ]
        }
      }
    `
  >
>
update: <
  path: <
    origin: "cli"
  >
  val: <
    ascii_val: "router bgp 15169"
  >
>
```

This transaction replaces the contents of the "openconfig" origins and then
updates the configuration with CLI content.  The first `replace` message
replaces the OpenConfig origin at the root (as specified by the zero-length
array of `PathElem` messages) to the specified JSON. The `update` operation
adds the CLI configuration of the device with with the string specified in the
`ascii_val` field. Per the gNMI specification, both operations MUST
successfully be applied, otherwise the configuration change should be rolled
back.

The expected resulting configuration will contain an enabled interface and a
router bgp 15169 statement.
