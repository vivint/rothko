// Copyright (C) 2017. See AUTHORS.

package main // import "github.com/spacemonkeygo/rothko/bin/rothko"

import (
	"context"
	"flag"
	"fmt"
	"os"
	"plugin"
	"strings"
	"sync"
	"time"

	"github.com/spacemonkeygo/rothko/accept"
	"github.com/spacemonkeygo/rothko/data"
	"github.com/spacemonkeygo/rothko/data/dists/tdigest"
	"github.com/spacemonkeygo/rothko/data/scribble"
	"github.com/spacemonkeygo/rothko/disk/files"
)

var (
	plugins = flag.String(
		"plugins",
		"",
		"comma separated list of plugins to load")

	acceptrixConfig = flag.String(
		"acceptrix",
		"",
		"comma separated list of acceptrix configs of the form name:config")

	// TODO(jeff): allow hooking into new kinds of distributions like the
	// acceptrix.

	compression = flag.Float64(
		"compression",
		5,
		"t-digest compression config value")

	filesDir     = flag.String("files.dir", "", "directory to store database")
	filesSize    = flag.Int("files.size", 512, "see files.Options")
	filesCap     = flag.Int("files.cap", 20480, "see files.Options")
	filesFiles   = flag.Int("files.files", 10, "see files.Options")
	filesBuffer  = flag.Int("files.buffer", 10000, "see files.Options")
	filesDrop    = flag.Bool("files.drop", false, "see files.Options")
	filesHandles = flag.Int("files.handles", 0, "see files.Options")
	filesWorkers = flag.Int("files.workers", 0, "see files.Options")

	periodicDump = flag.Duration(
		"periodic_dump",
		10*time.Minute,
		"how long between dumps of the scribbler into the database")
)

func main() {
	flag.Parse()
	if err := run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context) (err error) {
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

	// plugins should do all their registrations during their init function.
	for _, path := range strings.Split(*plugins, ",") {
		_, err := plugin.Open(path)
		if err != nil {
			return err
		}
	}

	// construct all the acceptrixes
	var acceptrixes []accept.Acceptrix
	for _, config := range strings.Split(*acceptrixConfig, ",") {
		parts := strings.SplitN(config, ":", 1)
		name, config := parts[0], ""
		if len(parts) > 1 {
			config = parts[1]
		}

		maker := accept.Lookup(parts[0])
		if maker == nil {
			return fmt.Errorf("unknown acceptrix: %q", name)
		}
		acceptrix, err := maker(ctx, config)
		if err != nil {
			return err
		}
		acceptrixes = append(acceptrixes, acceptrix)
	}

	// construct the scribbler and disk storage
	scr := scribble.NewScribbler(tdigest.Params{
		Compression: *compression,
	})

	fi := files.New(*filesDir, files.Options{
		Size:    *filesSize,
		Cap:     *filesCap,
		Files:   *filesFiles,
		Buffer:  *filesBuffer,
		Drop:    *filesDrop,
		Handles: *filesHandles,
		Workers: *filesWorkers,
	})

	// run all the acceptrixes on the scribbler
	errs := make(chan error, len(acceptrixes)+1) // ugh channels
	for _, acceptrix := range acceptrixes {
		acceptrix := acceptrix
		launch(func() { errs <- acceptrix.Run(ctx, scr) })
	}

	// launch the worker that periodically dumps in to the database
	launch(func() { errs <- periodicallyDump(ctx, scr, fi) })

	// launch the files database worker
	launch(func() { fi.Run(ctx) })

	// wait for an error i guess
	return <-errs
}

// TODO(jeff): this is too tightly coupled to the implementation details of
// the files database.

func periodicallyDump(ctx context.Context, scr *scribble.Scribbler,
	fi *files.DB) (err error) {

	bufs := sync.Pool{
		New: func() interface{} { return make([]byte, *filesSize) },
	}

	done := ctx.Done()
	ticker := time.NewTicker(*periodicDump)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return ctx.Err()

		case <-ticker.C:
			var err error
			scr.Capture(ctx, func(metric string, rec data.Record) bool {
				// check if we're cancelled
				select {
				case <-done:
					err = ctx.Err()
					return false
				default:
				}

				// marshal the record, reusing memory if possible
				data := bufs.Get().([]byte)
				if size := rec.Size(); cap(data) < size {
					data = make([]byte, size)
				} else {
					data = data[:size]
				}
				_, err = rec.MarshalTo(data)
				if err != nil {
					return false
				}

				// TODO(jeff): log the error that this returns
				fi.QueueCB(ctx, metric, rec.StartTime, rec.EndTime,
					data, func(written bool, err error) {
						// TODO(jeff): handle the input params appropriately?
						// probably just logging.

						bufs.Put(data)
					})
				return true
			})

			// errors "captured" from the closure are fatal.
			if err != nil {
				return err
			}
		}
	}
}
