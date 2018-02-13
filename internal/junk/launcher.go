// Copyright (C) 2018. See AUTHORS.

package junk

import (
	"context"
	"sync"
)

type Launcher struct {
	mu    sync.Mutex
	queue []func(ctx context.Context, errch chan error)
}

func (l *Launcher) Queue(fn func(ctx context.Context, errch chan error)) {
	l.mu.Lock()
	l.queue = append(l.queue, fn)
	l.mu.Unlock()
}

func (l *Launcher) Run(ctx context.Context) error {
	// steal the queue
	l.mu.Lock()
	queue := l.queue
	l.queue = nil
	l.mu.Unlock()

	// set it up so we wait for all the workers to exit
	var wg sync.WaitGroup
	wg.Add(len(queue))
	defer wg.Wait()

	// set up to cancel the workers when we get any error
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// make an error channel big enough for every worker to send in an error
	errch := make(chan error, len(queue))

	// off to the races!
	for _, fn := range queue {
		fn := fn
		go func() {
			defer wg.Done()
			fn(ctx, errch)
		}()
	}

	// wait for and return the first error
	return <-errch
}
