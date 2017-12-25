// Copyright (C) 2017. See AUTHORS.

package rothko

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/zeebo/errs"
)

// Main is the entrypoint of the rothko binary. Callers should only be main
// packages, and their main function should look like
//
//	func main() { rothko.Main() }
//
// It is this way so that it is easier to build custom binaries with plugins
// already imported.
func Main() {
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 || args[0] == "help" {
		printUsage(os.Stderr)
		fmt.Fprintln(os.Stderr)

		return
	}

	err := run(context.Background(), args)
	if err == nil {
		return
	}

	switch {
	case ErrInvalidParameters.Has(err):
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr)
		printUsage(os.Stderr)

	case ErrMissing.Has(err):
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr)
		listAvailable(os.Stderr)

	default:
		fmt.Fprintf(os.Stderr, "%+v\n", err)
	}

	fmt.Fprintln(os.Stderr)
	os.Exit(1)
}

func run(ctx context.Context, args []string) (err error) {
	just_list := false
	if args[0] == "list" {
		args, just_list = args[1:], true
	}

	config, _, err := ParseConfig(args)
	if err != nil {
		return err
	}

	// load all of the plugins
	if err := config.LoadPlugins(); err != nil {
		return err
	}

	if just_list {
		listAvailable(os.Stdout)
		fmt.Fprintln(os.Stdout)
		return nil
	}

	if config.Disk.Name == "" || config.Dist.Name == "" {
		return ErrInvalidParameters.New("must specify a disk and a dist")
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
	scr, err := config.LoadScribbler(ctx)
	if err != nil {
		return errs.Wrap(err)
	}

	// create the acceptrixes
	accs, err := config.LoadAcceptrixes(ctx)
	if err != nil {
		return errs.Wrap(err)
	}

	// create the disk
	di, err := config.LoadDisk(ctx)
	if err != nil {
		return errs.Wrap(err)
	}

	// keep track of the errors. the capacity is important to avoid deadlocks
	errch := make(chan error, len(accs)+2)

	// launch the acceptrixes
	for _, acc := range accs {
		acc := acc
		launch(func() { errch <- acc.Run(ctx, scr) })
	}

	// launch the worker that periodically dumps in to the database
	launch(func() { errch <- periodicallyDump(ctx, scr, di) })

	// launch the database worker
	launch(func() { errch <- di.Run(ctx) })

	// wait for an error i guess
	return <-errch
}
