// Copyright (C) 2018. See AUTHORS.

package rothko

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"plugin"

	"github.com/spacemonkeygo/rothko/api"
	"github.com/spacemonkeygo/rothko/config"
	"github.com/spacemonkeygo/rothko/data"
	_ "github.com/spacemonkeygo/rothko/database/files"
	_ "github.com/spacemonkeygo/rothko/dist/tdigest"
	"github.com/spacemonkeygo/rothko/dump"
	"github.com/spacemonkeygo/rothko/internal/junk"
	_ "github.com/spacemonkeygo/rothko/listener/graphite"
	"github.com/spacemonkeygo/rothko/registry"
	"github.com/zeebo/errs"
)

// Main is the entrypoint to any rothko binary. It is exposed so that it is
// easy to create custom binaries with your own enhancements.
func Main(conf config.Config) {
	started, err := run(context.Background(), conf)
	if err == nil {
		return
	}

	fmt.Fprintf(os.Stderr, "%+v\n", err)
	if !started {
		fmt.Fprintln(os.Stderr, "=== parsed config")
		conf.WriteTo(os.Stderr)
	}

	os.Exit(1)
}

// TODO(jeff): add logging with external about the start up process.

func run(ctx context.Context, conf config.Config) (started bool, err error) {
	// load the plugins
	for _, path := range conf.Main.Plugins {
		_, err := plugin.Open(path)
		if err != nil {
			return false, errs.Wrap(err)
		}
	}

	// create a launcher to keep track of all the tasks
	var launcher junk.Launcher

	// create the database
	db, err := registry.NewDatabase(ctx,
		conf.Database.Kind, conf.Database.Config)
	if err != nil {
		return false, errs.Wrap(err)
	}

	// create the distribution params from the registry
	dist_params, err := registry.NewDistribution(ctx,
		conf.Dist.Kind, conf.Dist.Config)
	if err != nil {
		return false, errs.Wrap(err)
	}

	// create the writer
	w := data.NewWriter(dist_params)

	// create and launch the listeners
	for _, entity := range conf.Listeners {
		listener, err := registry.NewListener(ctx, entity.Kind, entity.Config)
		if err != nil {
			return false, errs.Wrap(err)
		}
		launcher.Queue(func(ctx context.Context, errch chan error) {
			errch <- listener.Run(ctx, w)
		})
	}

	// create the dumper
	dumper := dump.New(dump.Options{
		DB:     db,
		Period: conf.Main.Duration,
	})

	// launch the worker that periodically dumps in to the database
	launcher.Queue(func(ctx context.Context, errch chan error) {
		errch <- dumper.Run(ctx, w)
	})

	// launch the database worker
	launcher.Queue(func(ctx context.Context, errch chan error) {
		errch <- db.Run(ctx)
	})

	// launch the api server
	launcher.Queue(func(ctx context.Context, errch chan error) {
		// TODO(jeff): basic auth
		// TODO(jeff): tls
		// TODO(jeff): proper CORS
		errch <- http.ListenAndServe(conf.API.Address, api.New(db))
	})

	// wait for an error
	return true, launcher.Run(ctx)
}
