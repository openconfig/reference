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

* Install the protoc compiler and any language specific generators.
 * [How to install](https://developers.google.com/protocol-buffers/docs/gotutorial)

* Compile the proto.
```
cd $GOPATH/src/github.com/openconfig/reference/rpc
protoc --go_out=plugins=grpc:. *.proto
```
