# Mixing Schemas in gNMI

**Contributors**: aashaikh, hines, robjs, csl, dloher
October 2016, Updated December 2022  
*Implementation Status*: Merged

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

### Transactionality of Sets with multiple Origins

Where a `SetRequest` specifies more than one `origin` - i.e., two or more
operations whose path include more than one origin - manipulations to all
affected trees MUST be considered as a single transaction. That is to say, only
if all transactions succeed should the `SetResponse` indicate success. If any
of the transactions fail, the contents of all origins MUST be rolled back, and
an error status returned upon responding to the `Set` RPC.

## Special considerations for a setRequest with openconfig and cli origins

### Use cases
The operational use cases for a setRequest `replace` operation containing both OC and CLI origins include:

1. The ability to support incremental transition from CLI based configuration to OpenConfig configuration.
2. A device may implement configuration items which are only defined in CLI while other configuration items are supported with OpenConfig.
3. The desire to use a full configuration replacement strategy.

To support these use cases a special meaning to a setRequest with replace origin `cli` and replace origin `openconfig` is defined.

### SetRequest behavior

A `SetRequest` message using the `replace` operation for both `cli` and `openconfig` origins SHOULD be performed as follows:

1. `origin` `cli` SHOULD appear first.  
2. The entire contents of the `ascii` encoded `cli` data SHOULD appear in a single update.
3. The `replace` operation for origin `cli` SHOULD be interpeted as a full configuration replacement.  In other words, the target device SHOULD erase the `cli` configuration and replace it with the contents of the `cli` update message.
4. `origin` `openconfig` should appear second.  
5. The target device should merge the `cli` and `openconfig` and apply the merged configuration as a single replace operation, replacing the combined `cli` and `openconfig` configuration of the device.
6. If an error occurs in merging or applying the merged configuration the device should behave as specified in [Transactionality of Sets with multiple Origins](#transactionality-of-sets-with-multiple-origins).

## Examples

### Example: OpenConfig and CLI origin `replace` operation

If a client wishes to replace both OpenConfig and CLI-modelled data
concurrently, it can send the following `SetRequest`:

```json
replace: <
  path: <
    origin: "cli"
  >
  val: <
    ascii_val: "
    qos traffic-class 0 name target-group-BE0
    qos tx-queue 0 name BE0"
  >
>
replace: <
  path: <
    origin: "openconfig"
  >
  val: <
    json_val: `
      {  
        "qos": {
          "classifiers": {
            "classifier": [
              {
                "config": {
                  "name": "dscp_based_classifier_ipv4",
                  "type": "IPV4"
                },
                "name": "dscp_based_classifier_ipv4",
                "terms": {
                  "term": [
                    {
                      "actions": {
                        "config": {
                          "target-group": "target-group-BE0"
                        }
                      },
                      "conditions": {
                        "ipv4": {
                          "config": {
                            "dscp-set": [
                              4, 5, 6, 7
                            ]
                          }
                        }
                      },
                      "config": {
                        "id": "1"
                      },
                      "id": "1"
                    },
                  ]
                }
              }
            ]
          },
          "forwarding-groups": {
            "forwarding-group": [
              {
                "config": {
                  "fabric-priority": 5,
                  "name": "target-group-BE0",
                  "output-queue": "BE0"
                },
                "name": "target-group-BE0"
              }
            ]
          },
          "interfaces": {
            "interface": [
              {
                "interface-id": "Port-Channel6",
                "output": {
                  "queues": {
                    "queue": [
                      {
                        "config": {
                          "name": "BE0",
                        },
                        "name": "BE0"
                      }
                    ]
                  },
                  "scheduler-policy": {
                    "config": {
                      "name": "scheduler"
                    }
                  }
                }
              }
            ]
          },
          "queues": {
            "queue": [
              {
                "config": {
                  "name": "BE0"
                },
                "name": "BE0"
              }
            ]
          },
          "scheduler-policies": {
            "scheduler-policy": [
              {
                "config": {
                  "name": "scheduler"
                },
                "name": "scheduler",
                "schedulers": {
                  "scheduler": [
                    {
                      "config": {
                        "sequence": 1
                      },
                      "inputs": {
                        "input": [
                          {
                            "config": {
                              "id": "BE0",
                              "input-type": "QUEUE",
                              "queue": "BE0"
                            },
                            "id": "BE0"
                          }
                        ]
                      },
                      "sequence": 1
                    }
                  ]
                }
              }
            ]
          }
        }
      }
    `
  >
>
```

This transaction replaces the entire contents of the device configuration with
a merger of the `openconfig` and `cli` origin data.  Both the `openconfig` and
`cli` data are intentionally simplified in this example, leaving out parts
required for real operation such as management network and interface
configuration necessary to maintain a connection to the device.

The `cli` `replace` operation replaces the full device configuration with the
string specified in the `ascii_val` field.  The `openconfig` `replace`
operation is performed at the root of the OpenConfig origin (as specified by
the zero-length array of `PathElem` messages) to the specified JSON.

The result is the BE0 queue is created in by `cli` and referenced by the
`openconfig` configuration.  

Per the gNMI specification, both `replace` operations MUST successfully be
applied, otherwise the entire configuration change should be rolled back.
