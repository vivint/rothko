// Copyright (C) 2017. See AUTHORS.

// package disk provides interfaces to disk storage of data.
package disk // import "github.com/spacemonkeygo/rothko/disk"

import (
	"context"
)

type Writer interface {
	Queue(ctx context.Context, metric string, start, end int64,
		data []byte) (err error)
}

type Source interface {
	Query(ctx context.Context, metric string, start, end int64) (
		Iterator, error)
	QueryLatest(ctx context.Context, metric string) ([]byte, error)

	Applications(ctx context.Context) (Iterator, error)
	Metrics(ctx context.Context, application string) (Iterator, error)
}

// Iterator is an iterator over a list of bytes or strings, like a
// https://golang.org/pkg/bufio/#Scanner.
type Iterator interface {
	Next(ctx context.Context) bool
	Bytes(ctx context.Context) []byte
	String(ctx context.Context) string
	Err(ctx context.Context) error
}
