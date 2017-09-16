// Copyright (C) 2017. See AUTHORS.

package files

import (
	"context"

	"github.com/spacemonkeygo/rothko/disk"
)

// Query returns an iterator over all of the data slices that overlap their
// start and end time with the provided values.
func (db *DB) Query(ctx context.Context, metric string, start int64,
	end int64) (disk.Iterator, error) {

	panic("not implemented")
}

// QueryLatest returns the latest value stored for the metric.
func (db *DB) QueryLatest(ctx context.Context, metric string) ([]byte, error) {
	panic("not implemented")
}

// Metrics returns an iterator over all of the metric names that are stored.
func (db *DB) Metrics(ctx context.Context) (disk.Iterator, error) {
	panic("not implemented")
}
