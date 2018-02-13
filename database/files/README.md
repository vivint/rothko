# package files

`import "github.com/spacemonkeygo/rothko/database/files"`

package files implements a disk.Source and disk.Writer

## Usage

```go
var Error = errs.Class("files")
```

#### type DB

```go
type DB struct {
}
```

DB is a database implementing database.Sink and database.Source using a file on
disk for each metric.

#### func  New

```go
func New(dir string, opts Options) *DB
```
New constructs a database with directory rooted at dir and the provided options.

#### func (*DB) Metrics

```go
func (db *DB) Metrics(ctx context.Context,
	cb func(name string) (bool, error)) (err error)
```
Metrics calls the callback once for every metric stored.

#### func (*DB) PopulateMetrics

```go
func (db *DB) PopulateMetrics(ctx context.Context) (err error)
```
PopulateMetrics walks the directory tree of the metrics recreating the in-memory
cache of metric names. It should be called periodically.

#### func (*DB) Query

```go
func (db *DB) Query(ctx context.Context, metric string, end int64,
	buf []byte, cb database.ResultCallback) error
```
Query calls the ResultCallback with all of the data slices that end strictly
before the provided end time in strictly decreasing order by their end. It will
continue to call the ResultCallback until it exhausts all of the records, or the
callback returns false.

#### func (*DB) QueryLatest

```go
func (db *DB) QueryLatest(ctx context.Context, metric string, buf []byte) (
	start, end int64, data []byte, err error)
```
QueryLatest returns the latest value stored for the metric. buf is used as
storage for the data slice if possible.

#### func (*DB) Queue

```go
func (db *DB) Queue(ctx context.Context, metric string, start int64,
	end int64, data []byte, cb func(bool, error)) (err error)
```
Queue adds the data for the metric and the given start and end times. If the
start time is before the last end time for the metric, no write will happen. The
callback is called with the error value of writing the metric.

#### func (*DB) Run

```go
func (db *DB) Run(ctx context.Context) error
```
Run will read values from the Queue and persist them to db. It returns when the
context is done.

#### type Options

```go
type Options struct {
	Size  int // size of each record
	Cap   int // cap of the number of records per file
	Files int // the number of historical files per metric

	Tuning Tuning // tuning parameters
}
```

Options is a set of options to configure a database.

#### type Tuning

```go
type Tuning struct {
	// Buffer controls the number of records that can be queued for writing.
	Buffer int

	// Drop, when true, will cause queued records to be discarded if the
	// buffer is full.
	Drop bool

	// Handles controls the number of open file handles for metrics in the
	// cache. If 0, then 1024 less than the soft limit of file handles as
	// reported by getrlimit will be used.
	Handles int

	// Workers controls the number of parallel workers draining queued values
	// into files. If zero, one less than GOMAXPROCS worker is used.
	//
	// The number of workers should be less than GOMAXPROCS, because each
	// worker deals with memory mapped files. The go runtime will not be able
	// to schedule around goroutines blocked on page faults, which could cause
	// goroutines to starve.
	Workers int
}
```

Tuning controls some tuning details of the database.
