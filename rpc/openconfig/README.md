# openconfig.proto

**Note: This package is deprecated - and has been replaced with the gRPC Network
Management Interface**. The gNMI specification can be found in 
[openconfig/reference](https://github.com/openconfig/reference/tree/master/rpc/gnmi),
and the protobuf service definition and reference code in
[openconfig/gnmi](https://github.com/openconfig/gnmi).

---

The openconfig package defines a [gRPC](http://www.grpc.io/) service for interacting with network devices
based on OpenConfig models.

This package and its contents are a *work-in-progress* and subject to change.  It is provided
as an example implementation of the
[OpenConfig RPC reference specification](available at github.com/openconfig/public/tree/master/release/models/rpc)
but also contains some additional capabilities not included in the base
RPC specification.

*Note: this is not an official Google product*

# Howto Generate the OpenConfig proto
* [Protocol buffer overview](https://developers.google.com/protocol-buffers)
* [Install protoc](https://developers.google.com/protocol-buffers/docs/proto3#generating)
 * Make sure you use the protoc version 3.0.0 currently this is only available on github.
 ```
 git clone https://github.com/google/protobuf
 ```
* Install any language specific generators and compile the proto.
 * Go based installation instructions
  * [How to compile proto for Go](https://developers.google.com/protocol-buffers/docs/gotutorial#compiling-your-protocol-buffers)
  * Compile the proto.
  ```
  cd $WORKSPACE
  protoc -I . github.com/openconfig/reference/rpc/openconfig/openconfig.proto  --go_out=plugins=grpc:.
  ```
 * C++ based installation instructions
  * [How to compile proto for C++](https://developers.google.com/protocol-buffers/docs/cpptutorial#compiling-your-protocol-buffers)
  * [Install gRPC plugin](https://github.com/grpc/grpc/blob/release-0_13/INSTALL.md)
  * Compile the proto.
  ```
  cd $WORKSPACE
  protoc -I ./ --grpc_out=. --plugin=protoc-gen-grpc=`which grpc_cpp_plugin` github.com/openconfig/reference/rpc/openconfig/openconfig.proto
  protoc -I ./ --cpp_out=. github.com/openconfig/reference/rpc/openconfig/openconfig.proto
  ```
 * Python based installation instructions
  * [How to compile proto for python](https://developers.google.com/protocol-buffers/docs/pythontutorial#compiling-your-protocol-buffers)
  * [Install gRPC plugin](https://github.com/grpc/grpc/blob/release-0_13/INSTALL.md)
  * Compile the proto.
  ```
  cd $WORKSPACE
  protoc -I. --python_out=. --grpc_out=. --plugin=protoc-gen-grpc=`which grpc_python_plugin` github.com/openconfig/reference/rpc/openconfig/openconfig.proto
  ```
