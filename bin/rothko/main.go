// Copyright (C) 2017. See AUTHORS.

package main // import "github.com/spacemonkeygo/rothko/bin/rothko"

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"plugin"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/spacemonkeygo/rothko/accept"
	"github.com/spacemonkeygo/rothko/data"
	_ "github.com/spacemonkeygo/rothko/data/dists/tdigest"
	"github.com/spacemonkeygo/rothko/data/scribble"
	"github.com/spacemonkeygo/rothko/disk"
	_ "github.com/spacemonkeygo/rothko/disk/files"
	"github.com/zeebo/errs"
)

var (
	InvalidParameters = errs.Class("invalid parameters")
	Missing           = errs.Class("missing")
)

func printUsage(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
usage: rothko [list|help] [parameters...]

parameters are of the form <kind>:<value> and there are four kinds:

	plugin:    pass "plugin:<path>" to load the plugin
	acceptrix: pass "acceptrix:<name>:<config>" to add an acceptrix
	disk:      pass "disk:<name>:<config>" to use the disk
	dist:      pass "dist:<name>:<config>" to use the distribution sketch

disk and dist are required. config may either be a string literal or a path to
a file containing the data.

the acceptrix is used to read data typically from a network interface and add
it to the disk using the distribution sketch. there may be multiple acceptrix
declarations.

for example:

	rothko \
		plugin:spacemonkey.so \
		acceptrix:sm/collector:0.0.0.0:9000 \
		disk:rothko/disk/files:files.json \
		dist:rothko/dist/tdigest:compression=5

will load the spacemonkey.so plugin, use the sm/collector acceptrix instructed
to listen on 0.0.0.0:9000, use the rothko files database configured from
files.json, and use the tdigest sketch with a compression of 5.

if you run "rothko list" and pass a set of plugins, the set of registered
acceptrixes, dists, and disks are outputted. run "rothko help" to see this
message.
`))
}

func listAvailable(w io.Writer) {
	tw := newTabbed(w)
	tw.write("kind", "name", "registrar")
	for _, reg := range accept.List() {
		tw.write("acceptrix", reg.Name, reg.Registrar)
	}
	for _, reg := range disk.List() {
		tw.write("disk", reg.Name, reg.Registrar)
	}
	for _, reg := range data.List() {
		tw.write("dist", reg.Name, reg.Registrar)
	}
	tw.flush()
}

func main() {
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
	case InvalidParameters.Has(err):
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr)
		printUsage(os.Stderr)

	case Missing.Has(err):
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
		listAvailable(os.Stderr)
		return nil
	}

	if config.Disk.Name == "" || config.Dist.Name == "" {
		return InvalidParameters.New("must specify a disk and a dist")
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
		return Missing.New("unknown dist: %q", config.Dist.Name)
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
			return Missing.New("unknown acceptrix: %q", name_config.Name)
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
		return Missing.New("unknown disk: %q", config.Dist.Name)
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

func periodicallyDump(ctx context.Context, scr *scribble.Scribbler,
	di disk.Disk) (err error) {

	// TODO(jeff): configs?

	bufs := sync.Pool{
		New: func() interface{} { return make([]byte, 1024) },
	}

	done := ctx.Done()
	ticker := time.NewTicker(10 * time.Minute)
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
				di.Queue(ctx, metric, rec.StartTime, rec.EndTime,
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

type tabbed struct {
	tw  *tabwriter.Writer
	err error
}

func newTabbed(w io.Writer) *tabbed {
	return &tabbed{
		tw: tabwriter.NewWriter(w, 0, 8, 3, ' ', 0),
	}
}

func (t *tabbed) write(values ...string) {
	if t.err == nil {
		_, t.err = fmt.Fprintln(t.tw, strings.Join(values, "\t"))
	}
}

func (t *tabbed) flush() {
	if t.err == nil {
		t.err = t.tw.Flush()
	}
}
