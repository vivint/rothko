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
}

// New constructs a database with directory rooted at path and the provided
// options.
func New(path string, opts Options) *DB {
	return &DB{
		path: path,
		opts: opts,
	}
}
