// Copyright (C) 2017. See AUTHORS.

// query is a command line interface to querying metrics.
package main // import "github.com/spacemonkeygo/rothko/bin/query"

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/spacemonkeygo/rothko/data"
	"github.com/spacemonkeygo/rothko/data/dists"
	"github.com/spacemonkeygo/rothko/disk/files"
	"github.com/spacemonkeygo/rothko/internal/junk"
	"github.com/zeebo/errs"
)

var invalidUsage = errs.Class("invalid usage")

var (
	filesConfigPath = flag.String("files_config", "files.json",
		"path to json file containing the config for the files backend")
)

func main() {
	flag.Parse()

	err := run(context.Background(), flag.Args())
	if err == nil {
		return
	}

	switch {
	case invalidUsage.Has(err):
		fmt.Fprintf(os.Stderr, "%v\n\n", err)
		printUsage(os.Stderr)

	default:
		fmt.Fprintf(os.Stderr, "%+v", err)
	}

	os.Exit(1)
}

func run(ctx context.Context, args []string) (err error) {
	if len(args) == 0 {
		return invalidUsage.New("no command specified")
	}

	files_config_data, err := ioutil.ReadFile(*filesConfigPath)
	if err != nil {
		return errs.Wrap(err)
	}

	var filesOptions struct {
		files.Options
		Dir string
	}

	if err := json.Unmarshal(files_config_data, &filesOptions); err != nil {
		return errs.Wrap(err)
	}

	di := files.New(filesOptions.Dir, filesOptions.Options)

	switch cmd := args[0]; cmd {
	case "latest":
		if len(args) == 1 {
			return errs.New("no metric specified")
		}

		start, end, data, err := di.QueryLatest(ctx, args[1], nil)
		if err != nil {
			return err
		}
		return printData(start, end, data)

	default:
		return invalidUsage.New("unknown command: %q", args[0])
	}

	return nil
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
usage: query <command> [args...]

command can be one of:

	latest:   query the latest value for the metric specified in args
`))
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
