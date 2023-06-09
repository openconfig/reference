# union_replace gNMI method

- [union\_replace gNMI method](#union_replace-gnmi-method)
  - [1. Objective](#1-objective)
  - [2. Definitions](#2-definitions)
    - [2.1 Network Operating System (NOS)](#21-network-operating-system-nos)
    - [2.2 OpenConfig Configuration (OC)](#22-openconfig-configuration-oc)
    - [2.3 Native YANG Configuration (NY)](#23-native-yang-configuration-ny)
    - [2.4 Native CLI Configuration (CLI)](#24-native-cli-configuration-cli)
    - [2.5 Startup configuration](#25-startup-configuration)
    - [2.6 Running or Active configuration](#26-running-or-active-configuration)
    - [2.7 Boot Config](#27-boot-config)
    - [2.8 bootz](#28-bootz)
    - [2.9 gNSI](#29-gnsi)
    - [2.10 Configuration item](#210-configuration-item)
    - [2.11 Full Device Configuration](#211-full-device-configuration)
    - [2.12 Dynamic configuration](#212-dynamic-configuration)
    - [2.13 Security configuration](#213-security-configuration)
    - [2.14 Bootstrap configuration](#214-bootstrap-configuration)
    - [2.15 Factory default configuration](#215-factory-default-configuration)
  - [3. Problem Statement](#3-problem-statement)
    - [3.1 The need for mixed origin configuration](#31-the-need-for-mixed-origin-configuration)
    - [3.2 What origins are to be mixed?](#32-what-origins-are-to-be-mixed)
  - [4. Device behavior for single-origin `SetRequests`](#4-device-behavior-for-single-origin-setrequests)
    - [4.1 Replace operation for Origin CLI  behavior](#41-replace-operation-for-origin-cli--behavior)
    - [4.2 Replace operation for path based Native Origin behavior](#42-replace-operation-for-path-based-native-origin-behavior)
  - [5 Device behavior for mixed native/CLI OC origin `SetRequests`](#5-device-behavior-for-mixed-nativecli-oc-origin-setrequests)
    - [5.1 Objective](#51-objective)
    - [5.2 Requirements for `SetRequest` with 2 origins](#52-requirements-for-setrequest-with-2-origins)
      - [5.2.2 union\_replace operation containing a native origin and origin OpenConfig](#522-union_replace-operation-containing-a-native-origin-and-origin-openconfig)
    - [5.3 Union behavior options](#53-union-behavior-options)
      - [5.3.1 Union behavior for CLI and OpenConfig](#531-union-behavior-for-cli-and-openconfig)
      - [5.3.2 Union behavior for Native and OpenConfig](#532-union-behavior-for-native-and-openconfig)
      - [5.3.3 Resolving issues with union between the origins](#533-resolving-issues-with-union-between-the-origins)
        - [5.3.3.1 Default values](#5331-default-values)
        - [5.3.3.2 Overlapping values in CLI and OC](#5332-overlapping-values-in-cli-and-oc)
        - [Option 1: Resolving issues with union of CLI and OC values with error](#option-1-resolving-issues-with-union-of-cli-and-oc-values-with-error)
        - [Option 2: Resolving issues with merging CLI and OC by overwriting](#option-2-resolving-issues-with-merging-cli-and-oc-by-overwriting)
      - [5.3.4 Overlapping values in NY and OC](#534-overlapping-values-in-ny-and-oc)
      - [5.3.5 Example of resolving default values](#535-example-of-resolving-default-values)
        - [Scenario A](#scenario-a)
        - [Scenario B](#scenario-b)
  - [6 Examples](#6-examples)
    - [6.1 Examples of expected configuration work flows](#61-examples-of-expected-configuration-work-flows)
    - [6.3 union\_replace example](#63-union_replace-example)
  - [7. gNSI, bootz and gNMI interoperation](#7-gnsi-bootz-and-gnmi-interoperation)
  - [8. References](#8-references)

## 1. Objective

There is an operational goal to ensure a network operating system’s (NOS)
complete operating configuration state is matched to the intended configuration
state.  A simple way to guarantee this is by generating the entire configuration
for the device and replacing whatever configuration the device currently has.
This declarative specification of the configuration is colloquially referred to
as ‘full config replace’. Although the concept is simple enough, there are
portions of the configuration which may not be modeled in OpenConfig.

## 2. Definitions

### 2.1 Network Operating System (NOS)

The operating system running on a network device which provides the management
interfaces.

### 2.2 OpenConfig Configuration (OC)

OC is defined as a set of OpenConfig working group approved YANG models. The
encoding for OC is defined in the gNMI specification.  The origin for a path in
gNMI is by default (empty) OC or may be explicitly stated as origin:openconfig.
Vendor augmentations to the OC models should be regarded as being part of
origin:openconfig.

### 2.3 Native YANG Configuration (NY)

NY is defined as a set of YANG models which are proprietary to a network
operating system (NOS). The origin of this configuration must be
`origin:<nos_native>`.  The NOS determines the value for `<nos_native>`. Example
values for `<nos_native>` might include:  `mynos_native` or `anothernos_native`
or any string which indicates the NOS name and the native designation.

### 2.4 Native CLI Configuration (CLI)

CLI configuration is defined as the NOS command line interface text used for
configuration.  In gNMI, this is encoded as ASCII text.  The format of the ASCII
text is agreed to out of band from  gNMI.  The origin of this configuration must
be `origin:<nos_cli>`.  Example values might include: `xros_cli`, `junos_cli`,
`srlinux_cli` or any string which indicates the NOS name and the CLI
designation.

### 2.5 Startup configuration

The configuration a device will load and use as the running configuration after
the base operating system has booted.

### 2.6 Running or Active configuration

The configuration which is in effect on the device.

### 2.7 Boot Config

The operating system (Linux) boot configuration which provides key-value data
when booting the kernel.  This is not the device startup configuration.  Boot
config items are managed by bootz.

### 2.8 bootz

[bootz](https://github.com/openconfig/bootz) describes structured data and API
for taking a device from factory defaults to a ready for production state.
bootz is intended to replace secure zero touch provisioning (sZTP) solutions.

Bootz defines partitions or namespaces for NOS configuration items and an order
for applying these namespaces to the NOS.  This is described briefly later in
this document and in detail at the [bootz](https://github.com/openconfig/bootz)
repository.

### 2.9 gNSI

 [gNSI](https://github.com/openconfig/gnsi) is referenced here as it allows for
 restricting access to OC and NY paths with it's pathz method.  In addition,
 when [bootz](https://github.com/openconfig/bootz#proposed-solution) is enabled
 the namespace controlled by gNSI must also be restricted from access by gNMI.

 Explicit access to gNSI restricted paths using a gNMI `SetRequest`, the
 `SetRequest` must be rejected with an error.

### 2.10 Configuration item

A configuration item is any attribute that may be assigned some value that is
then stored in a NOS that affects the behavior of the NOS or its supported
services.

### 2.11 Full Device Configuration

This term must be avoided. Because each NOS varies in which configuration items
are part of its native configuration, this term must not be used to describe the
intended configuration of a device. Examples of the variation in what is
included in a ‘full device configuration’  include NOS specific knobs, hardware
specific configuration, kernel boot information and security configuration
items.

### 2.12 Dynamic configuration

Configuration affected by a gNMI `SetRequest` is defined to be Dynamic
Configuration.  The term ‘Dynamic’ is used because gNMI configuration is
expected to be modified from time to time while the NOS is operationally in
service and forwarding traffic.

### 2.13 Security configuration

Security configuration is a special case of dynamic configuration and is managed
by gNSI.

### 2.14 Bootstrap configuration

This is the configuration necessary to boot and make a factory default NOS
manageable by gNMI.  This is inclusive of the ‘boot configuration’ defined
earlier, but also adds additional configuration needed to set initial security
login parameters, configure IP network connectivity for management purposes and
enable services such as gNMI, gNOI and gNSI.

### 2.15 Factory default configuration

The configuration that is present on the device when it is powered on for the
first time by the user.

## 3. Problem Statement

### 3.1 The need for mixed origin configuration

Defining the true, full configuration of network devices using OpenConfig may
not be feasible at some point in time for several reasons.  First, the features
to be configured may not have 100% coverage in OpenConfig (OC) models.  This may
be due to gaps in the OC modeling or gaps in vendor implementation of the OC
modeling.

Second, there are some configuration items which describe features that are
vendor proprietary and are not clearly in scope for today’s charter of
OpenConfig (e.g., vendor-specific hardware details). Third, there is an
operational use case to do emergency overrides of various config bits until
proper modeling and abstractions can be created.  This configuration must be
able to remove or replace other parts of the configuration and so must be
applied last in the configuration.  A phrase to capture this requirement is
"exception configlets”.

Therefore it is necessary to have some method to perform configuration using
both device native configuration and OpenConfig.

Note, there are portions of configuration that are necessary to preserve even in
a “full configuration replace” scenario.  Bootz defines the use cases and
approach for how to preserve configuration items, even when the configuration
items are expressed in different origins.

### 3.2 What origins are to be mixed?

There are two key use cases for mixing configuration schemas:

A network operator is transitioning from existing CLI-based configuration
generation to OpenConfig, and wants to achieve this incrementally. This, by
definition, requires mixing CLI text with OpenConfig modeled output. A network
operator wants to produce entirely modeled configuration, but needs a schema
other than OpenConfig to express some features, as discussed in the previous
section.

Thus, mixed schema must handle two combinations: 

- OC
- CLI OC and NY

Note, it is assumed that both the native schema and CLI model the entire
configuration surface area of the device. There is no requirement to support OC,
CLI & NY at the same time.

Furthermore note that these origins contain contents that overlap – which is the
motivation for the union_replace operation. Where origins are entirely disjoint,
they can be manipulated independently (whether in the same `SetRequest` or
different ones) since there is no interdependency between them.

## 4. Device behavior for single-origin `SetRequests`

### 4.1 Replace operation for Origin CLI  behavior

We first define a scenario to perform a replace of configuration items with only
CLI origin (not mixed origins).  Origin CLI does not support addressing a path
and is therefore interpreted as the empty path, `/` (ie: the “root” path) which
implies the entire contents of the CLI origin.  If bootz or gNSI are enabled
then the gNMI `SetRequest` MUST not replace any bootz or gNSI managed
configuration items.  In other words, the CLI origin replace only replaces the
configuration items in the dynamic config namespace.

When performing a gNMI `SetRequest` `replace` operation, the entire configuration
that the CLI  manages (subject to the namespace rules) should be ‘erased’ and
replaced with the contents of the CLI ASCII blob value.  It is acknowledged that
the entire CLI configuration may contain configuration items which overlap with
other origins.  CLI replace should only erase configuration that is writable by
CLI.  For example, certificates should not be erased by a `SetRequest` in a
scenario where installation of certificates is not managed by CLI.

This achieves a replacement of the dynamic CLI configuration without the need
for knowledge of what was already configured before.

A gNMI `SetRequest` `replace` operation using origin CLI must use only one
`replace` operation which contains the entire contents of the desired CLI
configuration.  One or more `update` operations may follow the `replace`
operation. The contents of the `update` are appended to the contents of that
specified in the `replace` operation.

Note, in a gNMI `SetRequest` the `delete`, `replace` and `update` operations are
performed in order as prescribed by the gNMI specification.  In origin CLI, the
delete operation is unsupported.

### 4.2 Replace operation for path based Native Origin behavior

A `replace` operation for path based native origins should be performed in the
same way as OpenConfig as defined in gNMI Specification section 3.4.

## 5 Device behavior for mixed native/CLI OC origin `SetRequests`

### 5.1 Objective

For this use case, the goal is to use OC wherever possible and a native origin
only where OC support is lacking.  The configuration generation software must
understand the device specific impact of native origin commands and their
interaction with OC, if any.

### 5.2 Requirements for `SetRequest` with 2 origins

Requirements are defined where 2 origins are present in a  `SetRequest`. Note, in
the below requirements, the use of “ `SetRequest`  operation” refers to the type of
operations used in a `SetRequest` message.

#### 5.2.2 union_replace operation containing a native origin and origin OpenConfig

A new operation named `union_replace` is introduced to meet the “full
configuration replace” goal using exactly one of NY or CLI, along with OC.

In this scenario, the target should replace the contents of its underlying
configuration using the union of the contents of the `union_replace` operation.
Union means the target must join the contents (e.g., the native and OpenConfig)
together to form a single configuration change which replaces the entire
configuration of the target.  If the target cannot join the contents together or
replace the configuration, the target MUST return an error.

An abstract set of steps a target may use to achieve `union_replace` behavior
are:

1. If using CLI origin, create a candidate configuration using factory defaults
   within the applicable namespace.  For example a NOS factory default
   configuration may assert that all ports, except a single management port are
   shutdown. The NOS may also assert that a DHCP client is enabled on the single
   management port and that protocols like BGP and ISIS are not instantiated and
   therefore are also not enabled.

When using union_replace, the user is expecting to assert all configuration
items needed to ensure the device has the intended state. For example, if a
device factory default configuration includes an admin account and the user
wants this removed, the user must include the negation of that user in the
union_replace.

If using NY origin, create a candidate configuration using the current running
configuration.  Because a NY origin has a path structure, the user can control
which portion of the tree should be replaced, including the entire NY tree.

1. Union the native data with the OC paths data and apply this to the candidate
   configuration.  The union behavior is described in 5.3 Union behavior
   options.

1. If an error is experienced when performing the union, the target responds to
   the RPC with a canonical error code of `INVALID_ARGUMENT` and SHOULD populate
   the error details with sufficient information for the union operation’s
   failure to be debugged.

1. Replace the running configuration with the candidate configuration.

1. If an error occurs when performing the configuration replacement the existing
   gNMI specification 3.4.3 applies.  The target rolls back all configuration
   changes.  The target responds to the RPC with a canonical error code of
   `INTERNAL` and SHOULD populate the error details with sufficient information
   for the replacement operation’s failure to be debugged.

### 5.3 Union behavior options

This section describes how the union between native and OC should be performed.
Different scenarios are described for handling CLI vs. Native origin union with
OpenConfig. In addition, there are two options for handling errors in the union
process. A NOS must assert which error handling option is implemented in its
documentation.

#### 5.3.1 Union behavior for CLI and OpenConfig

As the origins are joined together, the precedence specified (per origin) should
be followed to determine the order in which they are applied. The concept of
precedence is defined in bootz.

bootz defines CLI to be precedence 100, and `origin: openconfig` to be
precedence 110.  Configuration is merged in the reverse precedence order.  For
example,  precedence 0 should be merged last, overriding any higher precedence
value.  A Set operation specifying `union_replace` for both CLI and OC origins
should be performed as follows:

1. Create a candidate configuration using factory defaults.
     - Factory defaults are used because CLI is expected to represent the entire
      configuration, subject to any partitioning rules of bootz and gNSI.
2. Process OC paths, replacing the relevant configuration items in the candidate
   configuration.  This includes OC default values.
3. Process CLI items merging them to the candidate configuration.  Generate an
   error if there is conflict between OC and CLI as defined below.  Conflict and
   resolution is defined in
   [5.3.3](5.3.3-Resolving-issues-with-union-between-the-origins).

#### 5.3.2 Union behavior for Native and OpenConfig

We define a native origin to be precedence 100, and `origin: openconfig` to be
precedence 110.  A Set operation specifying `union_replace` operation for both
native and `openconfig` origins MUST be performed as follows. Additional
consideration is required because the native yang for a device is (by
definition) modeled and hence the path in the message must be considered.

1. Create a candidate config using the current running configuration.
   1. Running configuration is used because the NY paths may be a subtree.
2. Process OC paths, replacing the relevant subtree(s) in the candidate
   configuration.  This includes OC default values.
3. Process native paths, replacing the relevant subtree(s) in the candidate
   configuration.  Generate an error if there is conflict between OC and Native
   as defined below.

#### 5.3.3 Resolving issues with union between the origins

##### 5.3.3.1 Default values

The following rules govern how default values are applied when using two
origins.

1. Defaults from OC apply using the rules defined in the
   [RFC7950](https://www.rfc-editor.org/rfc/rfc7950#section-7.6.1) specification.
1. If the same configuration item is explicitly set by both origins to the same
   value, then that explicit value takes effect.
1. If a configuration item is explicitly set in one origin, but default in the
   other origin, then the explicit value always takes effect.
1. If a configuration item is not set and has only a default value in both
   origins then the OC default takes effect.
1. If a configuration item is explicitly set by both origins to two different
   values then it should be handled by one of the options in 5.3.5 below.

As a user of this specification, to ensure the desired value is applied, the
safest thing to do is to explicitly specify that value within the desired
origin's payload.

##### 5.3.3.2 Overlapping values in CLI and OC

Overlapping values refers to configuration items which are explicitly set by CLI
and OC.  Two options for how to resolve this are described.  Option 1 is
preferred and if supported should be indicated via documentation.  Option 2 is
an acceptable alternative if option 1 cannot be supported.

##### Option 1: Resolving issues with union of CLI and OC values with error

If a configuration item is explicitly set in CLI and OC using different values,
the target MUST return an `INVALID_ARGUMENT` error.  It may be necessary to set
the same values in both CLI and OC to conform to constraints, such as list keys,
which will need to be specified in both models if they are both targeting
configuration towards the same logical list entries.  It is up to the client to
provide CLI and OC which are not in conflict.

##### Option 2: Resolving issues with merging CLI and OC by overwriting

When applying CLI or NY data to the candidate configuration, overwrite any
configuration item already present following the order of merge precedence.

#### 5.3.4 Overlapping values in NY and OC

Where a pathed, NY origin configuration item overlaps with OC, the target MUST
return an `INVALID_ARGUMENT` error.  It is up to the client to provide the
native schema and OC which are not in conflict.

#### 5.3.5 Example of resolving default values

Here are two examples of when to use OC or CLI defaults in a  `SetRequest`.

##### Scenario A

ISIS is configured in OC, but the option for ISIS flex-algorithm which at the
moment does not exist in the OC schema, is configured using CLI.  In this case,
the OC defaults for ISIS should take effect for the OC paths used.  Flex-algo
should also be applied, using any relevant CLI defaults.

##### Scenario B

The BGP tree is configured using OC and ISIS is configured using CLI. In this
case, the OC defaults should be applied to BGP and the CLI defaults for ISIS
should apply.

Figure 1 - union_replace state machine

## 6 Examples

### 6.1 Examples of expected configuration work flows

1. Feature A is supported on CLI and configured via CLI
2. at a later time in a later s/w release, feature A becomes accessible via OC
3. configuration push omits the configuration in CLI and includes it in OC via
   union_replace.
4. The expected union_replace behavior is that the feature remains configured
   and no error is returned.  Abstracted from the client, the 'view' containing
   the feature is changed from CLI to OC.

### 6.3 union_replace example

```json
union_replace: {
  path: {
    origin: "acme_cli"
  }
  val: {
    ascii_val: "
 hostname myhost"
  }
}
union_replace: {
  path: {
    elem: {
      name: "network-instances"
    }
    elem: {
      name: "network-instance"
      key: {
        key: "name"
        value: "mgmtVrf"
      }
    }
    elem: {
      name: "interfaces"
    }
    elem: {
      name: "interface"
      key: {
        key: "id"
        value: "Management0"
      }
    }
  }
  val: {
    json_ietf_val:
      "{"
      "  \"openconfig-network-instance:config\": {"
      "    \"id\": \"Management0\""
      "  },"
      "  \"openconfig-network-instance:id\": \"Management0\""
      "}"
  }
}
```

## 7. gNSI, bootz and gNMI interoperation

Configuration affected by a gNMI `SetRequest` is defined to be Dynamic
Configuration.  THe term ‘Dynamic’ is used because this configuration is
expected to be modified while the NOS is running.   Security configuration is a
special case of dynamic configuration and is managed by gNSI.  Configuration
necessary to boot and make a factory default NOS manageable by gNMI is defined
as bootz configuration.

bootz configuration items can be expressed as OC, NY and CLI. Bootz defines an
order (or precedence) to apply the configuration name spaces to the NOS.  The
order is dynamic configuration, followed by bootz configuration, followed by
gNSI configuration.

When enabled, gNSI and bootz each reserve exclusive write access to the
configuration items they manage.  gNSI pathz may also be used to restrict read
access.  Since bootz specifies the use of native configuration, the target MUST
define, out of band, which configuration items are in scope of bootz which will
no longer be accessible to gNMI, regardless of origin.

When a gNMI replace or union_replace operation is performed, the gNSI and bootz
configuration items MUST not be affected.  If an openconfig path or `cli`
configuration item that is in scope of gNSI or bootz is explicitly referenced in
a gNMI  `SetRequest`, the `SetRequest` should be rejected with an error.  If this
behavior is not feasible, an acceptable alternative is to apply configuration in
the precedence order specified in bootz (OC -> NY -> CLI -> bootz -> gNSI).

## 8. References

* [gNMI specification](https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-specification.md)
* [bootz specification](https://github.com/openconfig/bootz)
* [gNSI specification](https://github.com/openconfig/gnsi)
