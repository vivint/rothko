# package data

`import "github.com/spacemonkeygo/rothko/data"`

package data provides types for handling rothko data.

    Package data is a generated protocol buffer package.

    It is generated from these files:
    	record.proto

    It has these top-level messages:
    	Record

## Usage

```go
var (
	ErrInvalidLengthRecord = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowRecord   = fmt.Errorf("proto: integer overflow")
)
```

#### type Record

```go
type Record struct {
	// start and end time in seconds since unix epoch utc
	StartTime int64 `protobuf:"varint,1,opt,name=start_time,json=startTime,proto3" json:"start_time,omitempty"`
	EndTime   int64 `protobuf:"varint,2,opt,name=end_time,json=endTime,proto3" json:"end_time,omitempty"`
	// the number of observations in the distribution
	Observations int64 `protobuf:"varint,3,opt,name=observations,proto3" json:"observations,omitempty"`
	// a serialized distribution. the kind tells us which type of distribution
	// it is.
	Distribution []byte `protobuf:"bytes,4,opt,name=distribution,proto3" json:"distribution,omitempty"`
	Kind         string `protobuf:"bytes,5,opt,name=kind,proto3" json:"kind,omitempty"`
	// minimum and maximum values observed
	Min   float64 `protobuf:"fixed64,6,opt,name=min,proto3" json:"min,omitempty"`
	Max   float64 `protobuf:"fixed64,7,opt,name=max,proto3" json:"max,omitempty"`
	MinId []byte  `protobuf:"bytes,8,opt,name=min_id,json=minId,proto3" json:"min_id,omitempty"`
	MaxId []byte  `protobuf:"bytes,9,opt,name=max_id,json=maxId,proto3" json:"max_id,omitempty"`
	// how many records have been merged into this.
	Merged int64 `protobuf:"varint,10,opt,name=merged,proto3" json:"merged,omitempty"`
}
```

Record is an observed distribution over some time period with some additional
data about observed minimums and maximums.

#### func (*Record) Descriptor

```go
func (*Record) Descriptor() ([]byte, []int)
```

#### func (*Record) Marshal

```go
func (m *Record) Marshal() (dAtA []byte, err error)
```

#### func (*Record) MarshalTo

```go
func (m *Record) MarshalTo(dAtA []byte) (int, error)
```

#### func (*Record) ProtoMessage

```go
func (*Record) ProtoMessage()
```

#### func (*Record) Reset

```go
func (m *Record) Reset()
```

#### func (*Record) Size

```go
func (m *Record) Size() (n int)
```

#### func (*Record) Unmarshal

```go
func (m *Record) Unmarshal(dAtA []byte) error
```

#### type Writer

```go
type Writer struct {
}
```

Writer keeps track of the distributions of a collection of metrics.

#### func  NewWriter

```go
func NewWriter(params dist.Params) *Writer
```
NewWriter makes a Writer that will return distributions using the associated
compression.

#### func (*Writer) Add

```go
func (s *Writer) Add(ctx context.Context, metric string,
	value float64, id []byte)
```
Add adds the metric value to the current set of records. It will be reflected in
the distribution of the records returned by Capture. WARNING: under some
concurrent scenarios, this can lose updates.

#### func (*Writer) Capture

```go
func (s *Writer) Capture(ctx context.Context,
	fn func(metric string, rec Record) bool)
```
Capture clears out current set of records for future Add calls and calls the
provided function with every record. You must not hold on to any fields of the
record after the callback returns.

#### func (*Writer) Iterate

```go
func (s *Writer) Iterate(ctx context.Context,
	fn func(metric string, rec Record) bool)
```
Iterate calls the provided function with every record. You must not hold on to
any fields of the record after the callback returns.
