// Copyright (C) 2017. See AUTHORS.

// query is a command line interface to querying metrics.
package main // import "github.com/spacemonkeygo/rothko/bin/query"

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spacemonkeygo/rothko"
	"github.com/spacemonkeygo/rothko/data"
	"github.com/spacemonkeygo/rothko/data/dists"
	"github.com/spacemonkeygo/rothko/disk"
	_ "github.com/spacemonkeygo/rothko/disk/files"
	"github.com/spacemonkeygo/rothko/internal/junk"
	"github.com/zeebo/errs"
)

func main() {
	err := run(context.Background())
	switch {
	case err == nil:
		return

	case rothko.ErrInvalidParameters.Has(err):
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr)
		printUsage(os.Stderr)

	case rothko.ErrMissing.Has(err):
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr)
		listAvailable(os.Stderr)

	default:
		fmt.Fprintf(os.Stderr, "%+v\n", err)
	}

	fmt.Fprintln(os.Stderr)
	os.Exit(1)
}

func run(ctx context.Context) (err error) {
	config, args, err := rothko.ParseConfig(os.Args[1:])
	if err != nil {
		return err
	}

	// load all of the plugins
	if err := config.LoadPlugins(); err != nil {
		return err
	}

	// load the disk
	di, err := config.LoadDisk(ctx)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return rothko.ErrInvalidParameters.New("no command specified")
	}

	switch cmd := args[0]; cmd {
	case "latest":
		if len(args) == 1 {
			return rothko.ErrInvalidParameters.New("no metric specified")
		}

		start, end, data, err := di.QueryLatest(ctx, args[1], nil)
		if err != nil {
			return err
		}
		return printData(start, end, data)

	default:
		return rothko.ErrInvalidParameters.New("unknown command: %q", args[0])
	}

	return nil
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
usage: query [parameters...] <command> [args...]

parameters are of the form <kind>:<value> and there are two kinds:

	plugin:    pass "plugin:<path>" to load the plugin
	disk:      pass "disk:<name>:<config>" to use the disk

disk is required. config may either be a string literal or a path to
a file containing the data.

command can be one of:

	latest:   query the latest value for the metric specified in args
`))
}

func listAvailable(w io.Writer) {
	tw := junk.NewTabbed(w)
	tw.Write("name", "registrar")
	for _, reg := range disk.List() {
		tw.Write(reg.Name, reg.Registrar)
	}
	tw.Flush()
}

func printData(start, end int64, buf []byte) error {
	var rec data.Record
	if err := rec.Unmarshal(buf); err != nil {
		return errs.Wrap(err)
	}
	dist, err := dists.Load(rec)
	if err != nil {
		return errs.Wrap(err)
	}

	tw := junk.NewTabbed(os.Stdout)
	tw.Write("start:", time.Unix(0, start).Format(time.RFC1123))
	tw.Write("end:", time.Unix(0, end).Format(time.RFC1123))
	tw.Write("obs:", fmt.Sprint(rec.Observations))
	tw.Write("kind:", rec.DistributionKind.String())
	tw.Write("data:", fmt.Sprintf("%x", rec.Distribution))
	tw.Write("min:", fmt.Sprint(rec.Min), fmt.Sprintf("%x", rec.MinId))
	tw.Write("max:", fmt.Sprint(rec.Max), fmt.Sprintf("%x", rec.MaxId))
	tw.Write("merged:", fmt.Sprint(rec.Merged))

	for x := 0.0; x <= 1.0; x += 1.0 / 32 {
		val := dist.Query(x)
		tw.Write(fmt.Sprintf("%0.2f:", x), fmt.Sprintf("%0.6f", val))
	}
	tw.Flush()

	return errs.Wrap(tw.Error())
}
