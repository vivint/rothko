// Copyright (C) 2018. See AUTHORS.

package rothko

import (
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"plugin"
	"time"

	"github.com/urfave/cli"
	"github.com/vivint/rothko/config"
	"github.com/vivint/rothko/data"
	"github.com/vivint/rothko/database"
	"github.com/vivint/rothko/dist"
	"github.com/vivint/rothko/external"
	"github.com/vivint/rothko/internal/junk"
	"github.com/vivint/rothko/registry"
	"github.com/zeebo/errs"
)

var demoCommand = cli.Command{
	Name:  "demo",
	Usage: "add some demo data to rothko",
	ArgsUsage: t(`
<path to rothko config>

To generate a rothko config, see the init command.
`),

	Description: t(`
The demo command fills your database with some data so you can play with it
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

		started, err := demo(context.Background(), conf)
		if started {
			return err
		}

		fmt.Printf("Invalid Configuration: %v\n", err)
		return handled.Wrap(err)
	},
}

func demo(ctx context.Context, conf *config.Config) (started bool, err error) {
	// load the plugins
	for _, path := range conf.Main.Plugins {
		external.Infow("loading plugin",
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
	external.Infow("creating database",
		"kind", conf.Database.Kind,
		"config", conf.Database.Config,
	)
	db, err := registry.NewDatabase(ctx,
		conf.Database.Kind, conf.Database.Config)
	if err != nil {
		return false, errs.Wrap(err)
	}

	// create the distribution params from the registry
	external.Infow("creating distribution",
		"kind", conf.Dist.Kind,
		"config", conf.Dist.Config,
	)
	params, err := registry.NewDistribution(ctx,
		conf.Dist.Kind, conf.Dist.Config)
	if err != nil {
		return false, errs.Wrap(err)
	}

	// queue the database
	launcher.Queue(func(ctx context.Context) error {
		external.Infow("starting database",
			"kind", conf.Database.Kind,
			"config", conf.Database.Config,
		)
		return db.Run(ctx)
	})

	// queue the demo data worker
	launcher.Queue(func(ctx context.Context) error {
		external.Infow("starting demo additions")
		return addDemoData(ctx, params, db)
	})

	// run our stuff
	return true, errs.Wrap(launcher.Run(ctx))
}

// addDemoData adds a bunch of random values to the database using the params
// to pick the kind of distribution.
func addDemoData(ctx context.Context, params dist.Params, db database.DB) (
	err error) {

	// random distributions to sample
	samplers := map[string]func(mean float64) float64{
		"normal": func(mean float64) float64 {
			return rand.NormFloat64() + mean
		},
		"exp": func(mean float64) float64 {
			return rand.ExpFloat64() * mean
		},
	}

	// deltas to the mean per time step
	mdeltas := map[string]float64{
		"increasing": 1,
		"stable":     0,
		"decreasing": -1,
	}

	const (
		duration     = 24 * time.Hour // how much history
		records      = 100            // how many records over the duration
		observations = 300            // number of observations per record
	)

	errors := make(chan error)
	now := time.Now()
	start := now.Add(-duration)
	tdelta := now.Sub(start) / records

	// loop over all the samplers and mean deltas.
	for sampler_name, sampler := range samplers {
		for mdelta_name, mdelta := range mdeltas {
			metric := fmt.Sprintf("%s.%s", sampler_name, mdelta_name)

			// start at the start and add the tdelta until we've passed now.
			t0, t1, mean := start, start.Add(tdelta), 0.0
			for t0.Before(now) {
				// create a record
				dist, err := params.New()
				if err != nil {
					return errs.Wrap(err)
				}

				min, max := math.Inf(1), math.Inf(-1)
				for i := 0; i < observations; i++ {
					val := sampler(mean)
					dist.Observe(val)
					if val < min {
						min = val
					}
					if val > max {
						max = val
					}
				}

				rec := data.Record{
					StartTime:    t0.UnixNano(),
					EndTime:      t1.UnixNano(),
					Observations: observations,
					Distribution: dist.Marshal(nil),
					Kind:         dist.Kind(),
					Min:          min,
					Max:          max,
					Merged:       1,
				}
				rec_data, err := rec.Marshal()
				if err != nil {
					return errs.Wrap(err)
				}

				// queue it and wait for it to be persisted
				err = db.Queue(ctx, metric, t0.UnixNano(), t1.UnixNano(),
					rec_data, func(written bool, err error) {
						errors <- err
					})
				if err != nil {
					return errs.Wrap(err)
				}
				if err := <-errors; err != nil {
					return errs.Wrap(err)
				}

				// step to the next time step and mean
				t0, t1, mean = t1, t1.Add(tdelta), mean+mdelta
			}
		}
	}

	return nil
}
