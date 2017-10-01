// Copyright (C) 2017. See AUTHORS.

package files

import (
	"context"

	"github.com/spacemonkeygo/rothko/disk"
)

// Query calls the ResultCallback with all of the data slices that overlap
// their start and end time with the provided values. The buf slice is
// used for storage of the data passed to the ResultCallback if possible.
// The data must not be modified, and no references must be kept after
// the callback returns.
func (db *DB) Query(ctx context.Context, metric string, start, end int64,
	buf []byte, cb disk.ResultCallback) error {

	db.locks.Lock(metric)
	defer db.locks.Unlock(metric)

	// acquire the datastructure encapsulating metric read logic
	met, err := db.newMetric(ctx, metric)
	if err != nil {
		return err
	}

	return met.Read(ctx, start, end, buf, cb)
}

// QueryLatest returns the latest value stored for the metric. buf is used
// as storage for the data slice if possible.
func (db *DB) QueryLatest(ctx context.Context, metric string, buf []byte) (
	start, end int64, data []byte, err error) {

	db.locks.Lock(metric)
	defer db.locks.Unlock(metric)

	// acquire the datastructure encapsulating metric read logic
	met, err := db.newMetric(ctx, metric)
	if err != nil {
		return 0, 0, nil, err
	}

	return met.ReadLast(ctx, buf)
}

// Metrics calls the callback once for every metric stored.
func (db *DB) Metrics(ctx context.Context, cb func(name string) error) (
	err error) {

	panic("not implemented")
}
