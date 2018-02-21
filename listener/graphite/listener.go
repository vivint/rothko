// Copyright (C) 2018. See AUTHORS.

package graphite

import (
	"context"

	"github.com/vivint/rothko/data"
)

// Listener implements the listener.Listener for the graphite wire protocol.
type Listener struct {
	address string
}

// New returns a Listener that when Run will listen on the provided address.
func New(address string) *Listener {
	return &Listener{
		address: address,
	}
}

// Run listens on the address and writes all of the metrics to the writer.
func (l *Listener) Run(ctx context.Context, w *data.Writer) (err error) {
	// TODO(jeff): do the thing
	<-ctx.Done()
	return nil
}
