#!/bin/bash

cp gnmi.proto gnmi.proto.orig
cat gnmi.proto | sed 's;github.com/golang/protobuf/ptypes/any/any.proto;google/protobuf/any.proto;g' > gnmi.proto.PY
cat gnmi.proto.PY | sed 's;github.com/google/protobuf/src/google/protobuf/descriptor.proto;google/protobuf/descriptor.proto;g' > gnmi.proto
python -m grpc.tools.protoc -I . --python_out=. --grpc_python_out=. gnmi.proto
rm gnmi.proto.PY
mv gnmi.proto.orig gnmi.proto
