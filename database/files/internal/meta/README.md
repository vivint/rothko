# package meta

`import "github.com/vivint/rothko/database/files/internal/meta"`



## Usage

```go
var (
	ErrInvalidLengthMeta = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowMeta   = fmt.Errorf("proto: integer overflow")
)
```

#### type Metadata

```go
type Metadata struct {
	// the alignment size of the file.
	Size_ int `protobuf:"varint,1,opt,name=size,proto3,casttype=int" json:"size,omitempty"`
	// points at the first available record.
	Head int `protobuf:"varint,2,opt,name=head,proto3,casttype=int" json:"head,omitempty"`
	// advisory start and end of the records in the file.
	Start       int64 `protobuf:"varint,3,opt,name=start,proto3" json:"start,omitempty"`
	End         int64 `protobuf:"varint,4,opt,name=end,proto3" json:"end,omitempty"`
	SmallestEnd int64 `protobuf:"varint,5,opt,name=smallest_end,json=smallestEnd,proto3" json:"smallest_end,omitempty"`
}
```

Metadata contains information about a file.

#### func (*Metadata) Marshal

```go
func (m *Metadata) Marshal() (dAtA []byte, err error)
```

#### func (*Metadata) MarshalTo

```go
func (m *Metadata) MarshalTo(dAtA []byte) (int, error)
```

#### func (*Metadata) Reset

```go
func (m *Metadata) Reset()
```

#### func (*Metadata) Size

```go
func (m *Metadata) Size() (n int)
```

#### func (*Metadata) Unmarshal

```go
func (m *Metadata) Unmarshal(dAtA []byte) error
```
