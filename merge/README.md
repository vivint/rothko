# package merge

`import "github.com/vivint/rothko/merge"`

package merge provides a merger for rothko data.

## Usage

```go
var Error = errs.Class("merge")
```
Error wraps all of the errors originating at this package.

#### func  Merge

```go
func Merge(ctx context.Context, opts MergeOptions) (
	out data.Record, err error)
```
Merge combines the records into one large record. The seed is used to do
deterministic merging.

#### type MergeOptions

```go
type MergeOptions struct {
	// Params are the parameters for the output distribution the merged record
	// should have.
	Params tdigest.Params

	// Records are the set of records to merge.
	Records []data.Record
}
```

MergeOptions are the arguments passed to Merge.

#### type Merger

```go
type Merger struct {
}
```

Merger allows iterative pushing of records in and constructs a series of merged
columns. The only requirement is that the end time on the records passed to push
are decreasing.

#### func  NewMerger

```go
func NewMerger(opts MergerOptions) *Merger
```
NewMerger constructs a Merger with the options.

#### func (*Merger) Finish

```go
func (m *Merger) Finish(ctx context.Context) ([]draw.Column, error)
```
Finish returns the set of columns to draw.

#### func (*Merger) Push

```go
func (m *Merger) Push(ctx context.Context, rec data.Record) error
```
Push adds the record to the Merger. The end time on the records passed to Push
must be decreasing.

#### func (*Merger) SetWidth

```go
func (m *Merger) SetWidth(width int)
```
SetWidth sets the width for all of the Push operations. Must be set before any
Push operations happen.

#### type MergerOptions

```go
type MergerOptions struct {
	Samples  int
	Now      int64
	Duration time.Duration
	Params   tdigest.Params
}
```

MergerOptions are the options the Merger needs to operate.
