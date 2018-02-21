// Copyright (C) 2018. See AUTHORS.

package junk

import (
	"context"
	"os"
	"os/signal"

	"github.com/vivint/rothko/external"
)

func WithSignal(ctx context.Context, sigs ...os.Signal) (
	context.Context, func()) {

	ctx, cancel := context.WithCancel(ctx)

	ch := make(chan os.Signal, len(sigs))
	signal.Notify(ch, sigs...)

	go func() {
		select {
		case <-ctx.Done():
		case sig := <-ch:
			external.Infow("signal received",
				"signal", sig.String(),
			)
			cancel()
			signal.Stop(ch)
		}
	}()

	return ctx, cancel
}
