# gNOI - gRPC Network Operations Interface


The gRPC Network Operation Interface, gNOI, defines a set of RPC services used for
invoking operational commands on network targets. These services are meant to be
used in conjunction with [gNMI](https://github.com/openconfig/reference/tree/master/rpc/gnmi)
for state management on network targets.  gNOI defines a number of services, most of which are
optional to implement on the target.  The gnoi.system.Service is the only service required to be
implemented for compliance.


**Note:** This is not an official Google product.
