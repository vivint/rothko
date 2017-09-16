// Copyright (C) 2017. See AUTHORS.

package files

import (
	"runtime"
	"sync"
	"syscall"
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
	// into files. If zero, the value of GOMAXPROCS at the start of the call
	// to New is used.
	Workers int
}

// DB is a database implementing disk.Writer and disk.Source using a file
// on disk for each metric.
type DB struct {
	path string
	opts Options

	// the queue of values and a sync.Pool containing byte slices since we want
	// to take ownership of the data passed in to Queue.
	queue chan queuedValue
	bufs  sync.Pool // contains []byte
	locks *lockPool

	// file handle cache for metrics
	mu   sync.Mutex
	toks map[string]cacheToken
	ch   *cache
}

// queuedValue represents some data queued to be written to disk.
type queuedValue struct {
	metric string
	start  int64
	end    int64
	data   []byte
	done   func(error)
}

// New constructs a database with directory rooted at path and the provided
// options.
func New(path string, opts Options) *DB {
	// set up the number of workers
	if opts.Workers == 0 {
		opts.Workers = runtime.GOMAXPROCS(-1)
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

	return &DB{
		path: path,
		opts: opts,

		queue: make(chan queuedValue, opts.Buffer),
		bufs: sync.Pool{
			New: func() interface{} { return make([]byte, opts.Size) },
		},
		locks: newLockPool(),

		toks: make(map[string]cacheToken),
		ch:   newCache(opts.Handles),
	}
}
