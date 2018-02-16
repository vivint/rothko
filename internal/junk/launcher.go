// Copyright (C) 2018. See AUTHORS.

package junk

import (
	"context"
	"sync"
)

type Launcher struct {
	mu    sync.Mutex
	queue []func(ctx context.Context) error
}

func (l *Launcher) Queue(fn func(ctx context.Context) error) {
	l.mu.Lock()
	l.queue = append(l.queue, fn)
	l.mu.Unlock()
}

func (l *Launcher) Run(ctx context.Context) error {
	l.mu.Lock()
	queue := l.queue
	l.queue = nil
	l.mu.Unlock()

	return Launch(ctx, queue...)
}

func Launch(ctx context.Context, tasks ...func(context.Context) error) error {
	var wg sync.WaitGroup
	wg.Add(len(tasks))
	defer wg.Wait()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errch := make(chan error, len(tasks))

	for _, fn := range tasks {
		fn := fn
		go func() {
			errch <- fn(ctx)
			wg.Done()
		}()
	}

	for range tasks {
		if err := <-errch; err != nil {
			return err
		}
	}
	return nil
}
