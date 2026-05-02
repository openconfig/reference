# gNMI - gRPC Network Management Interface

This repository contains the specification of the gRPC Network Management
Interface (gNMI). This service defines an interface
for a network management system to interact with a network element.

The protobuf specification is stored in
[openconfig/gnmi](https://github.com/openconfig/gnmi/tree/master/proto/gnmi).

The repository contents are as follows:

   * Specification for gNMI - [gnmi-specification.md](gnmi-specification.md).
   * Authentication Specification for gNMI - [gnmi-authentication.md](gnmi-authentication.md)
   * Path Conventions for gNMI - [gnmi-path-conventions.md](gnmi-path-conventions.md)
   * gNMI Support for Multiple Client Roles and Master Arbitration - [gnmi-master-arbitration.md](gnmi-master-arbitration.md)
   * gNMI/gNOI/SSH Dial-out via gRPC Tunnel - [gnmignoissh-dialout-grpctunnel.md](gnmignoissh-dialout-grpctunnel.md)

   * Deprecation of Decimal64 in gNMI - [decimal64-deprecation.md](decimal64-deprecation.md).
   * Representing gNMI Paths as Strings - [gnmi-path-strings.md](gnmi-path-strings.md).
   * `union_replace` gNMI method - [gnmi-union_replace.md](gnmi-union_replace.md).
   * Carrying Binary Wire Format Protobuf Messages in gNMI - [protobuf-vals.md](protobuf-vals.md).

gNMI Extensions:

   * Extensions to gNMI - [gnmi-extensions.md](gnmi-extensions.md).
   * gNMI Commit Confirmed Extension - [gnmi-commit-confirmed.md](gnmi-commit-confirmed.md).
   * gNMI Config Subscription Extension - [gnmi-config-subscriptions.md](gnmi-config-subscriptions.md).
   * gNMI Depth Extension - [gnmi-depth.md](gnmi-depth.md).
   * gNMI History Extension - [gnmi-history.md](gnmi-history.md).

**Note:** This is not an official Google product.

