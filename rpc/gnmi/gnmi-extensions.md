# Extensions to gNMI

**Contributors**: Rob Shakir (robjs@google.com), Carl Lebsack (csl@google.com),
Nick Ethier (nethier@jive.com), Anees Shaikh (aashaikh@google.com)

**Updated**: January 25th, 2018

**Version**: 0.1.0

## Extending gNMI

gNMI defines a standard set of RPCs which form the core protocol functionality.
In some implementations additional data is required within RPC payloads that is
not currently within the core protobuf definition. Extensions to gNMI define a
means to add new payload to gNMI RPCs for these use cases without requiring
changes in the core protocol specification.

An extension that is defined to a gNMI RPC MUST NOT modify the base behaviour of
the RPC. That is to say fields MUST NOT be interpreted in a manner which does
not match the base specification. The success or failure of an RPC MAY be
impacted by the extension. If the base specification's behaviour is to be
modified, an implementation MUST define a new service which specifies the
modified RPCs. A target MAY support both gNMI and such extension services as
different service endpoints on a common gRPC server.

gNMI extensions are defined to be carried within the `extension` field of each
top-level message of the gNMI RPCs. Extensions can added to both RPC request and
response messages, such that a client and target can both communicate extended
information. gNMI extensions are implemented directly within the request and
response messages to allow logging, debugging, or tracing frameworks to capture
their contents, and avoid the fragility of carrying extension information in
metadata.

## The Extension Message

The `Extension` message is defined within the `gnmi_ext.proto` specification. As
mentioned above, it is carried as a `repeated` field within each of the
top-level request and response gNMI messages.

## Well-Known and Registered Extensions

Well-known extensions are defined directly within the `gnmi_ext.proto` protocol
buffer. A well known extension is defined as one that is expected to be
supported by multiple implementations - for example, to support proxying, or
master arbitration between different writers.

Registered extensions are those that are more esoteric, or applicable to a
smaller set of use cases. In this case, the definition of the extension is
defined outside of the `gnmi.proto` or `gnmi_ext.proto` files. A registered
extension is given a unique identifier - which must be centrally registered in
the `gnmi_ext.proto` file. Registered extensions SHOULD provide a link to a
specification as to their operation. The gNMI `Extension` message is made
transparent to the definition of the protobuf message which defines the
extension by utilising a `bytes` field, which contains the binary marshalled
protobuf used to carry extension options.

Registered extensions MAY be promoted to well known extensions in the case that
their adoption becomes widespread.

## Extension Message Definition

The `Extension` message consists of a single `oneof` which may contain:

 * A well-known extension. Each well known extension defined in the
   `gnmi_ext.proto` file will be added to the `oneof` and assigned a unique
   field tag.
 * A registered extension, expressed as a `RegisteredExtension` message. The
   subfields of this message are:
   * An enumerated `id` field used to store the unique ID assigned to the
     registered extension.
   * A `bytes` field which stores the binary-marshalled protobuf for the
     extension.
