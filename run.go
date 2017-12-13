// Copyright (C) 2017. See AUTHORS.

package rothko

import (
	"context"
	"flag"
	"fmt"
	"os"
	"plugin"
	"sync"

	"github.com/spacemonkeygo/rothko/accept"
	"github.com/spacemonkeygo/rothko/data"
	"github.com/spacemonkeygo/rothko/data/scribble"
	"github.com/spacemonkeygo/rothko/disk"
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
	case errInvalidParameters.Has(err):
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr)
		printUsage(os.Stderr)

	case errMissing.Has(err):
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

	config, err := parseConfig(args)
	if err != nil {
		return err
	}

	// load all of the plugins
	for _, path := range config.Plugins {
		if _, err := plugin.Open(path); err != nil {
			return errs.Wrap(err)
		}
	}

	if just_list {
		listAvailable(os.Stdout)
		fmt.Fprintln(os.Stdout)
		return nil
	}

	if config.Disk.Name == "" || config.Dist.Name == "" {
		return errInvalidParameters.New("must specify a disk and a dist")
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

	// create the dist params and scribbler
	dist_maker := data.Lookup(config.Dist.Name)
	if dist_maker == nil {
		return errMissing.New("unknown dist: %q", config.Dist.Name)
	}
	params, err := dist_maker(ctx, config.Dist.Config)
	if err != nil {
		return errs.Wrap(err)
	}
	scr := scribble.NewScribbler(params)

	// create the acceptrixes
	var acceptrixes []accept.Acceptrix
	for _, name_config := range config.Acceptrixes {
		acceptrix_maker := accept.Lookup(name_config.Name)
		if acceptrix_maker == nil {
			return errMissing.New("unknown acceptrix: %q", name_config.Name)
		}
		acceptrix, err := acceptrix_maker(ctx, name_config.Config)
		if err != nil {
			return errs.Wrap(err)
		}
		acceptrixes = append(acceptrixes, acceptrix)
	}

	// create the disk
	disk_maker := disk.Lookup(config.Disk.Name)
	if disk_maker == nil {
		return errMissing.New("unknown disk: %q", config.Dist.Name)
	}
	di, err := disk_maker(ctx, config.Disk.Config)
	if err != nil {
		return errs.Wrap(err)
	}

	// keep track of the errors. the capacity is important to avoid deadlocks
	errch := make(chan error, len(config.Acceptrixes)+2)

	// launch the acceptrixes
	for _, acceptrix := range acceptrixes {
		acceptrix := acceptrix
		launch(func() { errch <- acceptrix.Run(ctx, scr) })
	}

	// launch the worker that periodically dumps in to the database
	launch(func() { errch <- periodicallyDump(ctx, scr, di) })

	// launch the database worker
	launch(func() { errch <- di.Run(ctx) })

	// wait for an error i guess
	return <-errch
}
