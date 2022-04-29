## Proposal to deprecate Decimal64 in gNMI

**Updated**: April 28, 2022
**Author**: Carl Lebsack

### Proposal

Deprecate and schedule removal of the Decimal64 message from gNMI and stipulate
all Decimal64 values in the OpenConfig models be delivered as
gNMI.TypedValue.double_val.

### Why is Decimal64 present today?

OpenConfig models are defined in [YANG](https://tools.ietf.org/html/rfc6020), a
modeling language which does not have a native IEEE-754 compliant binary
floating point type and has instead chosen to only
[support](https://tools.ietf.org/html/rfc6020#section-9.3) Decimal64 for real
number representation. Because of this choice, models developed within YANG,
including those of OpenConfig, have selected Decimal64 for data values where
fractional representation is necessary.

### Why is Decimal64 problematic for gNMI?

The primary problem is inefficiency which manifests in multiple ways by virtue
of a misalignment of the natural encoding of the real data and the decimal
encoding of the Decimal64 type. Before enumerating some of the inefficiencies in
practical use, we’ll start with a brief discussion of the ultimate source of the
data being modeled and the inappropriateness of Decimal64 for encoding.

All values within the OpenConfig models that currently use the YANG Decimal64 type are
fundamentally binary because they originate from
[Analog to Digital Converters](https://en.wikipedia.org/wiki/Analog-to-digital_converter)
(ADC) whereby a continuous analog signal, such as the laser power in an optical
transceiver, is discretized into a digital binary representation. This means
that despite the fact that we may render the outputs in a decimal form when
presenting a human readable output, the data itself is binary. It would likely have been preferable to use a native floating point typedef (such as the `ieeefloat32` typedef that was defined within `openconfig-types`) as opposed to `decimal64`. The continued use of `decimal64` is primarily for backwards-compatibility reasons.

Further, it is expected that a majority of clients using this data will
ultimately be using binary floating point representation and Decimal64 is simply
an awkward go-between induced by virtue of the YANG decision on only supporting
a single encoding for real types.  All time-series databases examined use binary
floating point for real numbers.  All major programming languages have native
support for binary floating point as do the vast majority of modern CPU hardware
where gNMI servers or its clients would likely run.

Decimal64, on the other hand, is “ intended for applications where it is
necessary to emulate decimal rounding exactly, such as financial and tax computations.”
([Wikipedia](https://en.wikipedia.org/wiki/Decimal64_floating-point_format)) The
decimal encoding is used for financial transactions because binary fractions
have no precise representations of many decimal values, e.g. 0.1, and therefore
can introduce error in computations where precise decimal rounding is required.

Where the source of modeled data originates from ADC hardware, which
themselves are subject to error and noise in their least significant bits, there
is no need for precise rounding even in binary, let alone decimal.
Additionally, often ADC are chosen based on cost with only as many bits as
necessary to differentiate a particular analog signal into a minimum number of
discrete values, often 16 bits or fewer. Given the lack of need for either the
precision, size or accuracy of Decimal64 to represent the underlying data we can
conclude that Decimal64 is in no way necessary and further enumerate some
inefficiencies in using this format in gNMI.

#### Poor infrastructure support for Decimal64

Because Decimal64 was introduced in 2008, there is very limited support for the
type natively in existing hardware, languages and libraries.  This is most
painful when integrating across systems and languages in places where gNMI is
designed to operate because the interchange format of Protocol Buffers has
[no native support of Decimal64](https://developers.google.com/protocol-buffers/docs/proto3#scalar).
The workaround is to create a [complex type](https://github.com/openconfig/gnmi/blob/master/proto/gnmi/gnmi.proto#L186)
using multiple integers as has been done in gNMI today.  This complex type
requires custom library code in every language to be used for encoding or
decoding these fractional values.  The code may not be complex but because the
protocol buffer message is bespoke to gNMI, it will not be common with any other
library code in existence for working with Decimal64 and thus has the potential
for individual implementers to introduce bugs that are not easily detected.

#### Unnecessary overhead of Decimal64

Additionally, in the context of streaming telemetry, the custom type adds
additional encoding and decoding steps to every value for every target server
and every client that handles the data. This propagates both the library
maintenance burden as well as inducing unnecessary computational overhead for
the repeated conversions on both servers and clients.

In addition to the redundant application level encoding and decoding overhead,
the representation itself is excessive for both in-memory and on the wire
transmission of the data.  As stated above, the source of the original data is
likely a low-bit-count ADC which can easily be encoded as a single 32bit binary
floating point number.  The message used in gNMI for Decimal64 uses two integer
values, one 32 bit the other 64 bit for a total of 96 bits, and creates a nested
structure which itself consumes additional space.  Given that most language
implementations of protocol buffers represent message types using pointers and
most systems in which gNMI will be deployed are likely 64-bit, this is an
additional 64-bit minimum for a total of 160 bits, or 5X overhead in storage.

Furthermore, because the Decimal64 type is a complex message, there is
additional computational overhead in the protocol buffer encoding and decoding
as compared to the native protocol buffer float type. A simple Go benchmark
(below) shows a 60% increase in protocol buffer encoding overhead.  This does
not include the previously mentioned application level encoding and decoding
between the original float and a Decimal64 message, which does not exist when a
native float type is used directly.

On an Intel Xeon W-2135 @ 3.7GHz

|                      | Double      | Decimal64   |
| BenchmarkEncode+Rand | 432.8 ns/op | 615.5 ns/op |
| BenchmarkRand        | 13.2 ns/op  | 12.6  ns/op |
| BenchmarkEncode only | 419.6 ns/op | 602.9 ns/op |

```go
package decimal64

import (
	"math/rand"
	"testing"

	"github.com/golang/protobuf/proto"

	pb "github.com/openconfig/gnmi/proto/gnmi/gnmi_go_proto"
)

func BenchmarkDoubleVal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		f := rand.Float64()
		t := &pb.TypedValue{Value: &pb.TypedValue_DoubleVal{DoubleVal: f}}
		_, err := proto.Marshal(t)
		if err != nil {
			b.Fatalf("marshal error: %v", err)
		}
	}
}

func BenchmarkRandFloat64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rand.Float64()
	}
}

func BenchmarkDecimal64Val(b *testing.B) {
	for i := 0; i < b.N; i++ {
		df := rand.Int63()
		t := &pb.TypedValue{Value: &pb.TypedValue_DecimalVal{&pb.Decimal64{Digits: df, Precision: 5}}}
		_, err := proto.Marshal(t)
		if err != nil {
			b.Fatalf("marshal error: %v", err)
		}
	}
}

func BenchmarkRandInt63(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rand.Int63()
	}
}
```

#### Confusing both implementers and operators

Another significant inefficiency of the use of Decimal64 is one of inter-
organizational human communication. The gNMI specification and Protocol Buffer
definition include two ways of rendering fractional values, Decimal64 and
double_val.  Because of this, we have some vendors using one or the other or both
methods for various paths.  Downstream clients need to handle both cases and
convert Decimal64 to float where it appears.  There is often the question of,
“What matches the official OpenConfig schema”.  The goal of this proposal is to
codify that answer as double_val for gNMI.
