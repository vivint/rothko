// Copyright (C) 2018. See AUTHORS.

package rothko

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"plugin"

	"github.com/spacemonkeygo/rothko/api"
	"github.com/spacemonkeygo/rothko/config"
	"github.com/spacemonkeygo/rothko/data"
	"github.com/spacemonkeygo/rothko/dump"
	"github.com/spacemonkeygo/rothko/external"
	"github.com/spacemonkeygo/rothko/internal/junk"
	"github.com/spacemonkeygo/rothko/registry"
	"github.com/urfave/cli"
	"github.com/zeebo/errs"
)

var runCommand = cli.Command{
	Name:  "run",
	Usage: "run rothko with some configuration",
	ArgsUsage: t(`
<path to rothko config>

To generate a rothko config, see the init command.
`),

	Description: t(`
The run command starts up the rothko system
`),

	Action: func(c *cli.Context) error {
		if err := checkArgs(c, 1); err != nil {
			return err
		}

		data, err := ioutil.ReadFile(c.Args().Get(0))
		if err != nil {
			return errs.Wrap(err)
		}

		conf, err := config.Load(data)
		if err != nil {
			return err
		}

		started, err := run(context.Background(), conf)
		if started {
			return err
		}

		fmt.Printf("Invalid Configuration: %v\n", err)
		return handled.Wrap(err)
	},
}

func run(ctx context.Context, conf *config.Config) (started bool, err error) {
	// load the plugins
	for _, path := range conf.Main.Plugins {
		external.Infow("loading plugin...",
			"plugin", path,
		)

		_, err := plugin.Open(path)
		if err != nil {
			return false, errs.Wrap(err)
		}
	}

	// create a launcher to keep track of all the tasks
	var launcher junk.Launcher

	// create the database
	external.Infow("creating database...",
		"kind", conf.Database.Kind,
		"config", conf.Database.Config,
	)
	db, err := registry.NewDatabase(ctx,
		conf.Database.Kind, conf.Database.Config)
	if err != nil {
		return false, errs.Wrap(err)
	}

	// create the distribution params from the registry
	external.Infow("creating distribution...",
		"kind", conf.Dist.Kind,
		"config", conf.Dist.Config,
	)
	dist_params, err := registry.NewDistribution(ctx,
		conf.Dist.Kind, conf.Dist.Config)
	if err != nil {
		return false, errs.Wrap(err)
	}

	// create the writer
	w := data.NewWriter(dist_params)

	// create and launch the listeners
	for _, entity := range conf.Listeners {
		entity := entity

		external.Infow("creating listener...",
			"kind", entity.Kind,
			"config", entity.Config,
		)
		listener, err := registry.NewListener(ctx, entity.Kind, entity.Config)
		if err != nil {
			return false, errs.Wrap(err)
		}

		launcher.Queue(func(ctx context.Context, errch chan error) {
			external.Infow("starting listener...",
				"kind", entity.Kind,
				"config", entity.Config,
			)
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
		external.Infow("starting dumper...")
		errch <- dumper.Run(ctx, w)
	})

	// launch the database worker
	launcher.Queue(func(ctx context.Context, errch chan error) {
		external.Infow("starting database...")
		errch <- db.Run(ctx)
	})

	// launch the api server
	launcher.Queue(func(ctx context.Context, errch chan error) {
		// TODO(jeff): basic auth
		// TODO(jeff): tls
		// TODO(jeff): proper CORS
		external.Infow("starting api...",
			"address", conf.API.Address,
		)
		errch <- http.ListenAndServe(conf.API.Address, api.New(db))
	})

	// wait for an error
	return true, launcher.Run(ctx)
}
