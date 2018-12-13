## Representing gNMI Paths as Strings

**Authors**: 
robjs<sup>⸸</sup>, aashaikh<sup>⸸</sup>, csl<sup>⸸</sup>, hines<sup>⸸</sup>.  
<sup>⸸</sup> @google.com

October 2018


## Overview

This document provides a specification for the representation of a gNMI
[`Path`](https://github.com/openconfig/gnmi/blob/master/proto/gnmi/gnmi.proto#L133)
message as a string. String encoding MUST NOT be used within the protocol, but
provides means to simplify human interaction with gNMI paths - for example, in
device or daemon configuration.


## String Representation of a Single Element

A gNMI path consists of a series of `PathElem` messages.  These messages
consist of a `name` and a map of `keys`. Each `PathElem` is represented as a
string in the form:

```
name[key1=val1][key2=val2]
```

Where `name` is the `PathElem` `name` field, `key1` represents a key of the
`keys` map, and `val1` represents the value corresponding to `key1` within the
map. Where multiple keys are present (as shown with `key1` and `key2`) the
values are encoded in the form:

```
[key=value]
```

and appended to the string. If the key value contains a `]` character, it is
escaped by preceding it with the backslash (`\`) character. Only `]` and `\`
characters must be escaped. Newlines and carriage returns are encoded as `\n`
and `\r` respectively.  Unicode characters are encoded as `\u1234` or `\U1234`.

If multiple keys exist for a particular element, they MUST be specified sorted
alphabetically by the key name.


## String Representation of an Entire Path

Using the format for each `PathElem` described above, the entire path is formed
by concatenating the series of strings with a `/` character between each
`PathElem`. If the path is absolute (i.e., starts at the schema root), it is
preceded by a `/` character. The `prefix + path` expressed in gNMI is always
absolute.


## Stringified Path Examples

<table>
  <tr>
   <td><strong>Path</strong>
   </td>
   <td><strong>String Encoding</strong>
   </td>
  </tr>
  <tr>
   <td>
<pre class="prettyprint">  prefix: &lt;
        elem: &lt;
                name: "a"
        >
  >
  path: &lt;
        elem: &lt;
                name: "b"
        >
        elem: &lt;
                name: "c"
        >
  ></pre>
   </td>
   <td><code>/a/b/c</code>
   </td>
  </tr>
  <tr>
   <td>
<pre class="prettyprint">&lt;
          elem: &lt;
                  name: "interfaces"
          >
            elem: &lt;
                  name: "interface"
                  key: &lt;
                          key: "name"
                          value: "Ethernet1/2/3"
                  >
          >
          elem: &lt;
                  name: "state"
          >
></pre>
   </td>
   <td><code>/interfaces/interface[name=Ethernet/1/2/3]/state</code>
   </td>
  </tr>
  <tr>
   <td>
<pre class="prettyprint">&lt;
                elem: &lt;
                        name: "interfaces"
                >
                elem: &lt;
                        name: "interface"
                        key: &lt;
                                key: "name"
                                value: "Ethernet1/2/3"
                        >
                >
                elem: &lt;
                        name: "state"
                >
                elem: &lt;
                        name: "counters"
                >
        >
></pre>
   </td>
   <td><code>/interfaces/interface[name=Ethernet/1/2/3]/state/counters</code>
   </td>
  </tr>
  <tr>
   <td>
<pre class="prettyprint">&lt;
        elem: &lt;
                name: "network-instances"
        >
        elem: &lt;
                name: "network-instance"
                key: &lt;
                        key: "name"
                        value: "DEFAULT"
                >
        >
        elem: &lt;
                name: "protocols"
        >
        elem: &lt;
                name: "protocol"
                key: &lt;
                        key: "identifier"
                        value: "ISIS"
                >
                key: &lt;
                        key: "name"
                        value: "65497"
                >
        >
></pre>
   </td>
   <td><code>/network-instances/network-instance[name=DEFAULT]/protocols/protocol[identifier=ISIS][name=65497]</code>
   </td>
  </tr>
  <tr>
   <td>
<pre class="prettyprint">&lt;
        elem: &lt;
                name: "foo"
                key: &lt;
                        key: "name"
                        value: "]"
                >
        >
></pre>
   </td>
   <td><code>/foo[name=\]]</code>
   </td>
  </tr>
  <tr>
   <td>
<pre class="prettyprint">&lt;
        elem: &lt;
                name: "foo"
                key: &lt;
                        key: "name"
                        value: "["
                >
        >
></pre>
   </td>
   <td><code>/foo[name=[]</code>
   </td>
  </tr>
  <tr>
   <td>
<pre class="prettyprint">&lt;
        elem: &lt;
                name: "foo"
                key: &lt;
                        key: "name"
                        value: "[\]"
                >
        >
></pre>
   </td>
   <td><code>/foo[name=[\\\]]</code>
   </td>
  </tr>
</table>


A reference implementation of path to string encoding, and string to path can
be found in ygot's
[pathstrings.go](https://github.com/openconfig/ygot/blob/master/ygot/pathstrings.go).

