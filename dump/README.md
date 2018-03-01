# package dump

`import "github.com/vivint/rothko/dump"`

package dump provides periodic dumping from a scribbler to disk.

## Usage

#### type Dumper

```go
type Dumper struct {
}
```

Dumper is a worker that periodically dumps from a Writer into a database.

#### func  New

```go
func New(opts Options) *Dumper
```
New constructs a Dumper with the given options.

#### func (*Dumper) Dump

```go
func (d *Dumper) Dump(ctx context.Context, w *data.Writer)
```
Dump writes all of the metrics Captured from the Writer into the DB associated
with the Dumper.

#### func (*Dumper) Run

```go
func (d *Dumper) Run(ctx context.Context, w *data.Writer) (err error)
```
Run dumps periodically, until the context is canceled. When the context is
canceled, it waits for any active Dump and returns.

#### type Options

```go
type Options struct {
	// The database to dump into.
	DB database.DB

	// How often to dump.
	Period time.Duration

	// How big a buffer to use for records. Defaults to 1024.
	Bufsize int
}
```

Options controls the options to the dumper.
