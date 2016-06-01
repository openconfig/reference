# telemetry

Reference implementations of a streaming telemetry agent and collector.
These are currently used for testing out real implementations but can be
used as the basis of a monitoring system based on streaming telemetry and
[OpenConfig](http://openconfig.net/) data models.

### Getting started

To build the collector, ensure you have go language tools installed
(available at [golang.org](golang.org/dl)) and that the `GOPATH`
environment variable is set to your Go workspace.

1. `go get github.com/openconfig/reference/telemetry
    * This will download the agent and collector code and dependencies into the src
subdirectory in your workspace.

2. `cd <workspace>/src/github.com/openconfig/reference/telemetry

3. `go build`

   * This will build the agent and collector binary and place it in the bin
subdirectory in your workspace.

### Contributing to telemetry

The telemetry tools are still a work-in-progress and we welcome contributions.  Please see
the `CONTRIBUTING` file for information about how to contribute to the codebase.

### Disclaimer

This is not an official Google product.
