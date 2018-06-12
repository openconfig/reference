# Carrying Binary Wire Format Protobuf Messages in gNMI
**Contributors**: robjs, csl  
May 2018  
*Implementation Status*: Proposal

## Problem Statement
For some applications, the data carried by gNMI is directly described by a
Protobuf IDL file which has been auto-generated from a YANG schema. Elements
within the Protobuf IDL are therefore 1:1 mapped to those within the YANG
schema - with the mapping between the YANG tree node and the protobuf
entity being established by the schema path of each.

When carrying such protobufs on the wire, it is undesirable to encode the
message name and namespace into a `protobuf.Any` field, as this:

 * Causes duplication of data on the wire - since the gNMI Notification's path
   (the concatenated `path` and `prefix` fields) already indicates which message
   is serialised within the Notification.

 * Results in incompatibilities when generated protobufs are within different
   namespaces, but represent the same YANG schema. This is typically the case
   when generating an augmented version of a standard schema. In this case, the
   augmented messages are backwards compatible with the standard messages, since
   they are a strict superset of the fields, and hence can be deserialised into
   the "standard" message. The differing type in `protobuf.Any` does not allow
   such deserialisation in standard library implementations.

This document describes how binary encoded protobufs are carried within gNMI,
and a mechanism for a schema-unaware client to retrieve the mapping of YANG path
to protobuf message name. This latter ability is of use where a schema-unaware
collector wishes at run-time to render paths to strings.

## Implementation

### Encoding in `TypedValue`

A new field is included within the `TypedValue` message's `value` `oneof` named
`protobuf_bytes`. When a target or client populates this field it indicates
that the content is a binary-marshalled Protobuf message. The protobuf
message's type is determined by the tuple of `<origin, Notification.prefix +
Update.path>`.

The contents of the field must be a byte array encoded according to the
[Protobuf Encoding
guide](https://developers.google.com/protocol-buffers/docs/encoding).

### Retrieving the Protobuf IDL

The target exposes a new origin `gnmi.schemas`, which is used to store schema
information relating to encoding in gNMI. This origin is used to store any YANG
model with an `openconfig-extension:origin` extension statement set to
`gnmi.schemas`.

The following YANG schema should be exposed under this origin.

```
module gnmi-protobuf-encoding {
  // Skip preamble

  container protobuf-typemap {
    list origin {
      key "name";

      leaf name {
        type string;
        description
          "The name of a origin for which protobuf
          message encoding is supported.";
      }

      list container {
        key "path";

        description
          "A list of containers within the YANG
          schema that has an associated protobuf
          message. The container is uniquely
          identified by its YANG schema path.";

        leaf path {
          type string;
          description
            "A YANG schema path - represented as
            each element separated by / characters.
            YANG schema paths MUST never contain
            keys.";
        }

        leaf message-name {
          type string;
          description
            "A Protobuf message name. This name
            should use the fully resolved path
            to the protobuf message.";
        }

        leaf augmented-message-name {
          type string;
          description
            "A Protobuf message name corresponding
            to the fully resolved path to a protobuf
            message that also includes augmentations
            made by the implementation.";
        }
      }
    }
  }
}
```

Within this data:

* The YANG schema path in the `path` leaf MUST represent the absolute schema
  tree ID to a particular container or list element within the YANG schema as
  defined by [RFC6020 Section
  6.5](https://tools.ietf.org/html/rfc6020#section-6.5). The schema path MUST
  NOT be a path to a leaf element.  
* The `message-name` leaf MUST represent a
  fully resolved Protobuf message name, in the form `source.tld/package.message`
  including any required hierarchy as described by the `type_url` field of
  [any.proto](https://github.com/google/protobuf/blob/master/src/google/protobuf/any.proto#L123-L149).
  The message name corresponds to the Protobuf message name used for that path,
  for example `proto.openconfig.net/openconfig.Interfaces`.
* The augmented message name is optionally populated by a device that also
  supports a protobuf message which contains additional fields added by local
  YANG modules which `augment` the base schema. This message MUST be backwards
  compatible with the message in specified in `message-name`. This message can
  be used by a client which wishes to unmarshal all fields in the message sent
  by the device.

<!-- TODO(robjs): link the augmentation document for protobuf -->

When a client issues a `Get` or `Subscribe` RPC to retrieve information from
the `gnmi.schemas` origin, the response MUST indicate the supported
encodings.


## Examples

#### Simple Mapping to Protobuf Schema

Assume that a target supports a module defining `/interfaces/interface/config`,
and its corresponding Protobuf encoding is the `Interfaces.Interface.Config`
message within an `proto.openconfig.net` package named `openconfig`. The target
should respond to a `GetRequest` specifying the `gnmi.schemas` origin with
the following content (shown in JSON encoding):

```
notification: <
  timestamp: 1257894000000000000
  update: <
    path: <
      origin: "gnmi.schemas"
    >
    val: <
      json_val: `
        {
          "protobuf-typemap": {
            "origin": [
              {
                "name": "openconfig",
                "container": [
                  {
                    "path": "/interfaces/interface/config",
                    "message-name": "proto.openconfig.net/openconfig.Interfaces.Interface.Config"
                  }
                ]
              }
            ]
          }
        }
      `
    >
  >
>
```

#### Mapping to a Standard and Augmented Schema

Suppose a target supports a standard YANG schema defining a schema path of
`/system`, along with a non-standard augmentation that adds a new set of nodes
to the `/system` container. In this case, the target should respond to a
`GetRequest` specifying the `gnmi.schemas` origin wtih the following content
(again, shown in JSON encoding):

```
notification: <
  timestamp: 1257894000000000000
  update: <
    path: <
      origin: "gnmi.schemas"
    >
    val: <
      json_val: `
        {
          "protobuf-typemap": {
            "origin": [
              {
                "name": "openconfig",
                "container": [
                  {
                    "path": "/system",
                    "message-name": "proto.openconfig.net/openconfig.System"
                    "augmented-message-name": "proto.vendorx.net/openconfig.System"
                  }
                ]
              }
            ]
          }
        }
      `
    >
  >
>
```
