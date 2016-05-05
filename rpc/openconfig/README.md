# openconfig.proto

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
* Install the protoc compiler and any language specific generators.
 * Go based installation instructions
  * [How to install protoc for Go](https://developers.google.com/protocol-buffers/docs/gotutorial)
  * Compile the proto.
  ```
  cd $GOPATH/src/github.com/openconfig/reference/rpc
  protoc --go_out=plugins=grpc:. *.proto
  ```
 * C++ based installation instructions
  * [How to install protoc for most languages](https://developers.google.com/protocol-buffers/docs/cpptutorial#compiling-your-protocol-buffers)
  * [Install gRPC plugin](https://github.com/grpc/grpc/blob/release-0_13/INSTALL.md)
  * Compile the proto.
  ```
  cd $GOPATH/src/github.com/openconfig/reference/rpc
  protoc -I ./ --grpc_out=. --plugin=protoc-gen-grpc=`which grpc_cpp_plugin` ./openconfig.proto
  protoc -I ./ --cpp_out=. ./openconfig.proto
  ```
 * Python based installation instructions
  * [How to install protoc for most languages](https://developers.google.com/protocol-buffers/docs/cpptutorial#compiling-your-protocol-buffers)
  * [Install gRPC plugin](https://github.com/grpc/grpc/blob/release-0_13/INSTALL.md)
  * Compile the proto.
  ```
  protoc -I. --python_out=. --grpc_out=. --plugin=protoc-gen-grpc=`which grpc_python_plugin` openconfig.proto
  ```
