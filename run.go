// Copyright (C) 2017. See AUTHORS.

package rothko

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/spacemonkeygo/rothko/data/scribble"
	"github.com/zeebo/errs"
)

func Main(options ...Option) {
	var opts Options
	for _, opt := range options {
		opt(&opts)
	}

	err := run(context.Background(), opts)
	if err == nil {
		return
	}

	fmt.Fprintf(os.Stderr, "%+v\n", err)
	os.Exit(1)
}

func run(ctx context.Context, opts Options) (err error) {
	if opts.Disk == nil || opts.DistParams == nil {
		return errs.New("must specify a disk and dist params in options")
	}

	// helpers to keep track of workers
	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	launch := func(fn func()) {
		wg.Add(1)
		go func() {
			fn()
			wg.Done()
		}()
	}

	// create the scribbler
	scr := scribble.NewScribbler(opts.DistParams)

	// keep track of the errors. the capacity is important to avoid deadlocks
	errch := make(chan error, len(opts.Acceptrixes)+2)

	// launch the acceptrixes
	for _, acc := range opts.Acceptrixes {
		acc := acc
		launch(func() { errch <- acc.Run(ctx, scr) })
	}

	// launch the worker that periodically dumps in to the database
	if opts.Dumper != nil {
		launch(func() { errch <- opts.Dumper.Run(ctx, scr) })
	}

	// launch the database worker
	launch(func() { errch <- opts.Disk.Run(ctx) })

	// wait for an error
	return <-errch
}
