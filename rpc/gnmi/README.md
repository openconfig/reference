# gNMI - gRPC Network Management Interface

This directory contains a reference
[protobuf3](https://github.com/google/protobuf/) service definition for the
gRPC Network Management Interface (gNMI). This service defines an interface
for a network management system to interact with a network element.

The repository contents are as follows:
 * Reference protobuf - [gnmi.proto](gnmi.proto).
 * Specification for gNMI - [gnmi-specification.md](gnmi-specification.md).
   * PDF of specification document
     [gnmi-specification.pdf](gnmi-specification.pdf)
   * Authentication Specification for gNMI - [gnmi-authentication.md](gnmi-authentication.md)
   * Path Conventions for gNMI - [gnmi-path-conventions.md](gnmi-path-conventions.md)
 * Generated Go code for gNMI - [gnmi.pb.go](gnmi.pb.go)
 * Generated Python code for gNMI - [gnmi_pb2.py](gnmi_pb2.py)

**Note:** This is not an official Google product.

This proto has external dependencies on `google/protobuf/any.proto` and
`google/protobuf/descriptor.proto`, which can be imported directly (or via
GitHub paths).
