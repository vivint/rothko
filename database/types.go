// Copyright (C) 2018. See AUTHORS.

package database

import (
	"context"
)

// Sink represents something that can add data about metrics.
type Sink interface {
	// Queue adds the data for the metric and the given start and end times. If
	// the start time is before the last end time for the metric, no write is
	// guaranteed to happen. The data is not required to be persisted to disk
	// after the call returns, and may be flushed asynchronously. If the cb
	// parameter is not nil, it will be called when the data has been handled.
	// Written indicates if the data was written to disk, and err is not nil
	// if there were errors.
	Queue(ctx context.Context, metric string, start, end int64,
		data []byte, cb func(written bool, err error)) (err error)
}

// ResultCallback is a function used to pass results back from Query. The data
// slice must not be modified, and no references must be kept after the
// function returns. Return if you will continue iterating.
type ResultCallback func(ctx context.Context, start, end int64, data []byte) (
	bool, error)

// Source can be used to read data about metrics.
type Source interface {
	// Query calls the ResultCallback with all of the data slices that end
	// strictly before the provided end time in strictly decreasing order by
	// their end. It will continue to call the ResultCallback until it exhausts
	// all of the records, or the callback returns false.
	Query(ctx context.Context, metric string, end int64, buf []byte,
		cb ResultCallback) error

	// QueryLatest returns the latest value stored for the metric. buf is used
	// as storage for the data slice if possible.
	QueryLatest(ctx context.Context, metric string, buf []byte) (
		start, end int64, data []byte, err error)

	// Metrics calls the callback once for every metric stored.
	Metrics(ctx context.Context, cb func(name string) (bool, error)) error
}

// DB represents a Source and a Sink.
type DB interface {
	Source
	Sink

	// Run will be called so that the DB can do asynchronous tasks. The
	// context will be canceled when it is expected to shut down.
	Run(ctx context.Context) error
}
