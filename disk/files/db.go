// Copyright (C) 2017. See AUTHORS.

package files

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/spacemonkeygo/rothko/disk"
	"github.com/spacemonkeygo/rothko/external"
)

// Options is a set of options to configure a database.
type Options struct {
	// The following three values determine the maximum amount of data that
	// will be stored per metric, as well as their retention policies given
	// how often records are sent in. The amount of space the database uses
	// should be bounded by:
	//
	//	number of metrics * size * cap * (files + 1)
	//
	// Assuming that size is large enough to typically hold the size of a
	// serialized record, you should be able to have historical data for a
	// period bounded by:
	//
	//	avg time between metric writes * size * cap * (files + 1)
	//
	// You will typically want to tune cap and files so that the typical access
	// patterns for reads of data fall entirely inside of one file. For
	// example, you would not want to have them chosen to only hold 12 hours
	// of data if you expect most queries to be over a 1 day period.

	Size  int // size of each record
	Cap   int // cap of the number of records per file
	Files int // the number of historical files per metric

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

	// Resources to use during operation.
	Resources external.Resources `json:"-"`
}

// DB is a database implementing disk.Writer and disk.Source using a file
// on disk for each metric.
type DB struct {
	dir  string
	opts Options

	// the queue of values and a sync.Pool containing byte slices since we want
	// to take ownership of the data passed in to Queue.
	queue chan queuedValue
	bufs  sync.Pool // contains []byte
	locks *lockPool

	// file handle cache for metrics
	fch *fileCache

	// cache of metric names. one per worker with unions on reads
	names_w_mu []sync.Mutex
	names_w    []map[string]struct{}

	// metric names cache for easy atomic swapping. is a map[string]struct{}
	// and it is readonly. the names_mu is held during population of the
	// value.
	names_mu sync.Mutex
	names    atomic.Value
}

var (
	// type assert the interfaces we expect to implement
	_ disk.Source = (*DB)(nil)
	_ disk.Sink   = (*DB)(nil)
	_ disk.Disk   = (*DB)(nil)
)

// queuedValue represents some data queued to be written to disk.
type queuedValue struct {
	metric string
	start  int64
	end    int64
	data   []byte
	done   func(bool, error)
}

// New constructs a database with directory rooted at dir and the provided
// options.
func New(dir string, opts Options) *DB {
	// set up the number of workers
	if opts.Workers == 0 {
		opts.Workers = runtime.GOMAXPROCS(-1) - 1

		// in the worst case, run one worker anyway
		if opts.Workers <= 0 {
			opts.Workers = 1
		}
	}

	// set up the number of handles
	if opts.Handles == 0 {
		var lim syscall.Rlimit
		if syscall.Getrlimit(syscall.RLIMIT_NOFILE, &lim) == nil {
			if int64(int(lim.Cur)) == int64(lim.Cur) {
				opts.Handles = int(lim.Cur) - 512
			}
		}
	}
	if opts.Handles < 0 {
		opts.Handles = 0
	}

	names_w := make([]map[string]struct{}, opts.Workers)
	for i := range names_w {
		names_w[i] = make(map[string]struct{})
	}

	return &DB{
		dir:  dir,
		opts: opts,

		queue: make(chan queuedValue, opts.Buffer),
		bufs: sync.Pool{
			New: func() interface{} { return make([]byte, opts.Size) },
		},
		locks: newLockPool(),

		fch: newFileCache(fileCacheOptions{
			Handles: opts.Handles,
			Size:    opts.Size,
			Cap:     opts.Cap,
		}),

		names_w_mu: make([]sync.Mutex, opts.Workers),
		names_w:    names_w,
	}
}

// newMetric constructs a *metric value for the database.
func (db *DB) newMetric(ctx context.Context, name string) (*metric, error) {
	return newMetric(ctx, metricOptions{
		fch:  db.fch,
		dir:  db.dir,
		name: name,
		max:  db.opts.Files,
		ext:  db.opts.Resources,
	})
}

// Run will read values from the Queue and persist them to disk. It returns
// when the context is done.
func (db *DB) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	wg.Add(db.opts.Workers)
	for i := 0; i < db.opts.Workers; i++ {
		go func(i int) {
			db.worker(ctx, i)
			wg.Done()
		}(i)
	}

	// wait for the workers who will exit when the context is done.
	wg.Wait()

	return ctx.Err()
}
