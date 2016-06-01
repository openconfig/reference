# cli
cli is a simple OpenConfig client implementation.  It can be used for
testing out an OpenConfig server implementation.

cli connects to a provided target using a query and will either perform a
Get or Subscribe (-use_subscribe) and return the results.  If -tls is not
specified the connection will be unsecure.
The basic per RPC authentication is provided by user/password authentication.
If -outfile is not provided the notifications will be output to stdout.

# Examples

## Get

Get performs a single OpenConfig Get request for the provided query paths. The paths must be comma separated and "/" must be used as a path separator.

```
./cli -target=127.0.0.1:10162 -tls -user=test -password=test -query=/bgp/global -outfile=/tmp/foo

./cli -target=127.0.0.1:10162 -tls -user=test -password=test -query=/interfaces/interface[name=Ethernet1/2/3]/state
```

## Subscribe

Subscribe performs an OpenConfig Subscribe request for the provided query paths. The paths must be comma separated and "/" must be used as a path separator. If -subscribe_once is provided the subscription will be closed once a SyncResponse message has been returned by the server.

```
./cli -target=127.0.0.1:10162 -tls -user=test -password=test -query=/interfaces/interface[name=*]/state/counters -use_subscribe -outfile=/tmp/foo -subscribe_once
```

# Setup
* Install gRPC

```
cd $GOPATH
go get google.golang.org/grpc
```

* From your $GOPATH/src:

```
cd $GOPATH/src
mkdir -p github.com/openconfig
cd github.com/openconfig
```

* Clone the reference proto into your $GOPATH

```
git clone https://github.com/openconfig/reference.git
```

* This will prompt you for your username and password

* Get the basic proto types installed correctly in your go path
 * you must have protoc already installed. [How to install](https://developers.google.com/protocol-buffers/docs/gotutorial)
 * Note: on linux boxes protoc installs to /usr/local/include however for
 OSX it is /usr/local/Cellar/protobuf/3.0.0-beta-2/include

```
cd $GOPATH/src
DIR=/usr/local/include;protoc -I $DIR $DIR/google/protobuf/any.proto
$DIR/google/protobuf/source_context.proto $DIR/google/protobuf/type.proto
--go_out=.
```

* (optional) Regenerate the openconfig.pb.go code

```
cd $GOPATH/github.com/openconfig/reference/rpc
protoc --go_out=plugins=grpc:. *.proto
```

* Build the cli

```
go build github.com/openconfig/reference/telemetry/collector/cli
```
