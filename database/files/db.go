// Copyright (C) 2018. See AUTHORS.

package files

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/vivint/rothko/database"
	"github.com/vivint/rothko/database/files/internal/sset"
	"github.com/vivint/rothko/external"
	"github.com/vivint/rothko/internal/junk"
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

	Tuning Tuning // tuning parameters
}

// Tuning controls some tuning details of the database.
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

// DB is a database implementing database.Sink and database.Source using a file
// on disk for each metric.
type DB struct {
	dir  string
	opts Options

	// the queue of values and a sync.Pool containing byte slices since we want
	// to take ownership of the data passed in to Queue.
	queue atomic.Value // contains chan queuedValue
	bufs  sync.Pool    // contains []byte
	locks *lockPool

	// file handle cache for metrics
	fch *fileCache

	// cache of metric names. one per worker with unions on reads
	names_w_mu []sync.Mutex
	names_w    []*sset.Set

	// metric names cache for easy atomic swapping. is a map[string]struct{}
	// and it is readonly. the names_mu is held during population of the
	// value.
	names_mu sync.Mutex
	names    atomic.Value

	// ensures that we only have one Run call
	running junk.Flag
}

var (
	// type assert the interfaces we expect to implement
	_ database.Source = (*DB)(nil)
	_ database.Sink   = (*DB)(nil)
	_ database.DB     = (*DB)(nil)
)

// queuedValue represents some data queued to be written to db.
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
	if opts.Tuning.Workers == 0 {
		opts.Tuning.Workers = runtime.GOMAXPROCS(-1) - 1

		// in the worst case, run one worker anyway
		if opts.Tuning.Workers <= 0 {
			opts.Tuning.Workers = 1
		}
	}

	// set up the number of handles
	if opts.Tuning.Handles == 0 {
		var lim syscall.Rlimit
		if syscall.Getrlimit(syscall.RLIMIT_NOFILE, &lim) == nil {
			if int64(int(lim.Cur)) == int64(lim.Cur) {
				opts.Tuning.Handles = int(lim.Cur) - 512
			}
		}
	}
	if opts.Tuning.Handles < 0 {
		opts.Tuning.Handles = 0
	}

	var queue atomic.Value
	queue.Store(make(chan queuedValue, opts.Tuning.Buffer))

	names_w := make([]*sset.Set, opts.Tuning.Workers)
	for i := range names_w {
		names_w[i] = sset.New(0)
	}

	return &DB{
		dir:  dir,
		opts: opts,

		queue: queue,
		bufs: sync.Pool{
			New: func() interface{} { return make([]byte, opts.Size) },
		},
		locks: newLockPool(),

		fch: newFileCache(fileCacheOptions{
			Handles: opts.Tuning.Handles,
			Size:    opts.Size,
			Cap:     opts.Cap,
		}),

		names_w_mu: make([]sync.Mutex, opts.Tuning.Workers),
		names_w:    names_w,
	}
}

// newMetric constructs a *metric value for the database.
func (db *DB) newMetric(ctx context.Context, name string, read_only bool) (
	*metric, error) {

	return newMetric(ctx, metricOptions{
		fch:  db.fch,
		dir:  db.dir,
		name: name,
		max:  db.opts.Files,
		ro:   read_only,
	})
}

// Run will read values from the Queue and persist them to db. It returns
// when the context is done.
func (db *DB) Run(ctx context.Context) error {
	// ensure only one active Run call.
	if err := db.running.Start(); err != nil {
		return Error.Wrap(err)
	}
	defer db.running.Stop()

	// load up the current queue to run on
	queue := db.queue.Load().(chan queuedValue)

	// queue up the workers
	var launcher junk.Launcher

	for i := 0; i < db.opts.Tuning.Workers; i++ {
		i := i
		launcher.Queue(func(ctx context.Context) error {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()

			db.worker(ctx, i, queue)
			return nil
		})
	}

	// queue up populating the metric names
	launcher.Queue(func(ctx context.Context) error {
		external.Infow("caching metric names")

		n := time.Now()
		err := db.PopulateMetrics(ctx)

		external.Infow("cached metric names",
			"duration", time.Since(n),
		)
		if err != nil {
			external.Errorw("caching metric names",
				"error", err.Error(),
			)
		}

		<-ctx.Done()
		return nil
	})

	// launch and wait for them
	err := launcher.Run(ctx)

	// clear out any cached files
	db.fch.Close()

	// set a new queue for the next calls to Queue
	db.queue.Store(make(chan queuedValue, db.opts.Tuning.Buffer))

	// close and empty the old queue. close is safe because all of the callers
	// writing to it recover any panics.
	close(queue)
	for val := range queue {
		db.bufs.Put(val.data)
		if val.done != nil {
			val.done(false, nil)
		}
	}

	// return any error from launching
	return err
}
