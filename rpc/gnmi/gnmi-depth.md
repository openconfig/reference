# gNMI Depth Extension

**Contributors:** Roman Dodin, Matthew MacDonald

**Date:** March 9, 2024

**Version:** 0.1.0

## 1. Purpose

The implicit recursive data fetching used by the gNMI servers makes the implementations simple. Whilst the implementation simplicity is important, the lack of server-side filtering for the data requested in gNMI RPCs may be considered a limitation for some gNMI users and systems.

The gNMI Depth Extension allows a client to control the depth of the recursion when the server evaluates a group of paths in the Subscribe or Get RPC.

Orchestration, Network Management and Monitoring Systems can benefit from this extension as it:

1. reduces the load on the server when data is to be fetched from the Network OS during the recursive data extraction
2. reduces the bytes on the wire payload, by sending fewer data

gNMI Depth Extension proto specification is defined in [gnmi_ext.proto](https://github.com/openconfig/gnmi/blob/master/proto/gnmi_ext/gnmi_ext.proto).

## 2. Demo model

To explain the concept of the depth-based filtering, consider the following model that is used in the implementation examples throughout this document:

```yang
    container basket {
      leaf name {
        type string;
      }

      leaf-list contents {
        type string;
      }

      list fruits {
        key "name";

        leaf name {
          type string;
        }

        leaf-list colors {
          type string;
        }

        leaf size {
          type string;
        }

        container origin {
          leaf country {
            type string;
          }
          
          leaf city {
            type string;
          }
        }
      }

      container description {
        leaf fabric {
          type string;
        }
      
      }
      container broken {
        presence "This container is broken";
        leaf reason {
          type string;
        }
      }
    }
```

It's tree representation:

```
module: app
  +--rw basket
     +--rw name?          string
     +--rw contents*      string
     +--rw fruits* [name]
     |  +--rw name      string
     |  +--rw colors*   string
     |  +--rw size?     string
     |  +--rw origin
     |     +--rw country?   string
     |     +--rw city?      string
     +--rw description
     |  +--rw fabric?   string
     +--rw broken!
        +--rw reason?   string
```

We populate this data schema with the following values:

```
    basket {
        contents [
            fruits
            vegetables
        ]
        fruits apples {
            size XL
            colors [
                red
                yellow
            ]
            origin {
                country NL
                city Amsterdam
            }
        }
        fruits orange {
            size M
        }
        description {
            fabric cotton
        }
        broken {
            reason "too heavy"
        }
    }
```

## 3. Concepts

The Depth extension allows clients to specify the depth of the subtree to be returned in the response. The depth is specified as the number of levels below the specified path.

The extension itself has a single field that controls the depth level:

```proto
message Depth {
  uint32 level = 1;
}
```

### 3.1 Depth level values

#### 3.1.1 Value 0

Depth value of 0 means no depth limit and behaves the same as if the extension was not specified at all.

#### 3.1.2 Value 1

Value of 1 means only the specified path and its direct children will be returned. See Children section for more info.

#### 3.1.2 Value of N+

Value of N+ where N>1 means all elements of the specified path up to N level and direct children of N-th level.

### 3.2 Children nodes

The Depth extension operates the value of "direct children of a schema node". What we understand by direct children:

1. leafs
2. leaf-lists

Only these elements are to be returned if depth extension with non-0 value is specified for a specified depth level.

### 3.3 RPC support

The Depth extension applies to Get and Subscribe requests only. When used with Capability and Set RPC the server should return an error.

## 4 Examples

Using the data model from Section 2 we will run through a set of examples using the patched version of [openconfig/gnmic](https://gnmic.openconfig.net/) client with the added Depth extension support. We can provide the patched gnmic binary for Linux x86_84 if you want to try it out.

### 4.1 depth 1, path `/basket`

The most common way to use the depth extension (as we see it) is to use it with level=1. This gets you the immediate child nodes of the schema node targeted by a path.

Consider the following gnmic command targeting `/basket` path:

```bash
gnmic -e json_ietf get --path /basket --depth 1
```

```json
[
  {
    "contents": [
      "fruits",
      "vegetables"
    ]
  }
]
```

As per the design, only the leaf and leaf-list nodes are returned. Since our `/basket` container has only `leaf-list` elements (no leafs) a single element `contents` is returned.

You can see how this makes it possible to reduce the amount of data extracted by the server and sent over the wire. Many applications might require fetching only leaf values of a certain container to make some informed decision without requiring any of the nested data.

### 4.2 depth 1, path `/basket/fruits`

When the path targets the list schema node, all elements of this list is returned with their children nodes

```bash
gnmic -e json_ietf get --path /basket/fruits --depth 1
```

```json
[
  {
    "fruits": [
      {
        "colors": [
          "red",
          "yellow"
        ],
        "name": "apples",
        "size": "XL"
      },
      {
        "name": "orange",
        "size": "M"
      }
    ]
  }
]
```

Again, please keep in mind that only leafs and leaf-lists are returned for every list element.

### 4.3 depth 2, path `/basket`

When the depth level is set to values >1, all elements from the path to the provided level value are returned in full with the last level including only leafs and leaf-lists.

```bash
gnmic -e json_ietf get --path /basket --depth 2
```

```json
[
  {
    "broken": {
      "reason": "too heavy"
    },
    "contents": [
      "fruits",
      "vegetables"
    ],
    "description": {
      "fabric": "cotton"
    },
    "fruits": [
      {
        "colors": [
          "red",
          "yellow"
        ],
        "name": "apples",
        "size": "XL"
      },
      {
        "name": "orange",
        "size": "M"
      }
    ]
  }
]
```

Here is what happens:

<img width="640" alt="image" src="https://github.com/openconfig/gnmi/assets/5679861/331a20a5-ff19-4932-873a-7aff9b36dd62">

The 1st level elements are returned, since depth level is 2.
On the 2nd level we return only leafs and leaf-lists, hence the `.fruits.origin` is not present.
