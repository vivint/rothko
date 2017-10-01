// Copyright (C) 2017. See AUTHORS.

// package disk provides interfaces to disk storage of data.
package disk // import "github.com/spacemonkeygo/rothko/disk"

import (
	"context"
)

// Sink represents something that can add data about metrics.
type Sink interface {
	// Queue adds the data for the metric and the given start and end times. If
	// the start time is before the last end time for the metric, no write is
	// guaranteed to happen. The data is not required to be persisted to disk
	// after the call returns, and may be flushed asynchronously.
	Queue(ctx context.Context, metric string, start, end int64,
		data []byte) (err error)
}

// SinkCB is an optional stronger interface than Sink.
type SinkCB interface {
	// QueueCB is the same as the Sink.Queue method except it can be passed a
	// callback that is called when the metric has been handled. Written
	// indicates if the data was written to disk, and err is not nil if there
	// were errors.
	QueueCB(ctx context.Context, metric string, start, end int64,
		data []byte, cb func(written bool, err error)) (err error)
}

// ResultCallback is a function used to pass results back from Query. The data
// slice must not be modified, and no references must be kept after the
// function returns.
type ResultCallback func(start, end int64, data []byte) error

// Source can be used to read data about metrics.
type Source interface {
	// Query calls the ResultCallback with all of the data slices that overlap
	// their start and end time with the provided values. The buf slice is
	// used for storage of the data passed to the ResultCallback if possible.
	// The data must not be modified, and no references must be kept after
	// the callback returns.
	Query(ctx context.Context, metric string, start, end int64, buf []byte,
		cb ResultCallback) error

	// QueryLatest returns the latest value stored for the metric. buf is used
	// as storage for the data slice if possible.
	QueryLatest(ctx context.Context, metric string, buf []byte) (
		start, end int64, data []byte, err error)

	// Metrics calls the callback once for every metric stored.
	Metrics(ctx context.Context, cb func(name string) error) error
}
