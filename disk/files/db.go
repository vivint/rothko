// Copyright (C) 2017. See AUTHORS.

package files

// DB is a database implementing disk.Writer and disk.Source using a file
// on disk for each metric.
type DB struct {
	path string
	opts Options
}

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
}

// New constructs a database with directory rooted at path and the provided
// options.
func New(path string, opts Options) *DB {
	return &DB{
		path: path,
		opts: opts,
	}
}
