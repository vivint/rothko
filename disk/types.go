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
	// after the call returns.
	Queue(ctx context.Context, metric string, start, end int64,
		data []byte) (err error)
}

// Source can be used to read data about metrics.
type Source interface {
	// Query returns an iterator over all of the data slices that overlap their
	// start and end time with the provided values.
	Query(ctx context.Context, metric string, start, end int64) (
		Iterator, error)

	// QueryLatest returns the latest value stored for the metric.
	QueryLatest(ctx context.Context, metric string) ([]byte, error)

	// Metrics returns an iterator over all of the metric names that are
	// stored.
	Metrics(ctx context.Context) (Iterator, error)
}

// Iterator is an iterator over a list of bytes or strings, like a
// https://golang.org/pkg/bufio/#Scanner.
type Iterator interface {
	Next(ctx context.Context) bool
	Bytes(ctx context.Context) []byte
	String(ctx context.Context) string
	Err(ctx context.Context) error
}
