# dump
--
    import "github.com/spacemonkeygo/rothko/dump"

package dump provides periodic dumping from a scribbler to disk.

## Usage

#### type Dumper

```go
type Dumper struct {
}
```


#### func  New

```go
func New(opts Options) *Dumper
```

#### func (*Dumper) Run

```go
func (d *Dumper) Run(ctx context.Context, w *data.Writer) (err error)
```

#### type Options

```go
type Options struct {
	DB      database.DB
	Period  time.Duration
	Bufsize int
}
```
