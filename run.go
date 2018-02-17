// Copyright (C) 2018. See AUTHORS.

package rothko

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"plugin"
	"syscall"
	"time"

	"github.com/spacemonkeygo/rothko/api"
	"github.com/spacemonkeygo/rothko/config"
	"github.com/spacemonkeygo/rothko/data"
	"github.com/spacemonkeygo/rothko/dump"
	"github.com/spacemonkeygo/rothko/external"
	"github.com/spacemonkeygo/rothko/internal/junk"
	"github.com/spacemonkeygo/rothko/internal/tgzfs"
	"github.com/spacemonkeygo/rothko/internal/tmplfs"
	"github.com/spacemonkeygo/rothko/registry"
	"github.com/spacemonkeygo/rothko/ui"
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
	// TODO(jeff): the following is terrible and has at least these problems:
	// 1. if the database Run call errors, calls to Queue can still proceed.
	//    this can lead at least to deadlocks, but it's still terrible that
	//    we can call stuff on the DB while it isn't Running.
	// 2. i have no idea if it's right
	// 3. the errors can't possibly be propagating right.

	// we have a very complicated shutdown order to allow us to attempt to get
	// one final dump in before exiting. first, we shut down everything except
	// the database, and then we call Dump on the Dumper with a context that
	// cancels in 60 seconds. Then we can shut down the database and wait for
	// all of the stuff to clean up.

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

		external.Infow("creating listener",
			"kind", entity.Kind,
			"config", entity.Config,
		)
		listener, err := registry.NewListener(ctx, entity.Kind, entity.Config)
		if err != nil {
			return false, errs.Wrap(err)
		}

		launcher.Queue(func(ctx context.Context) error {
			external.Infow("starting listener",
				"kind", entity.Kind,
				"config", entity.Config,
			)
			return listener.Run(ctx, w)
		})
	}

	// create the dumper
	dumper := dump.New(dump.Options{
		DB:     db,
		Period: conf.Main.Duration,
	})

	// create the api server
	//
	// TODO(jeff): basic auth
	// TODO(jeff): tls
	// TODO(jeff): proper CORS
	external.Infow("creating api")
	fs, err := tgzfs.New(ui.Tarball)
	if err != nil {
		return false, errs.Wrap(err)
	}
	srv := &http.Server{
		Addr:    conf.API.Address,
		Handler: api.New(db, tmplfs.New(fs)),
	}

	// launch the worker that periodically dumps in to the database
	launcher.Queue(func(ctx context.Context) error {
		external.Infow("starting dumper")
		return dumper.Run(ctx, w)
	})

	// launch the api server
	launcher.Queue(func(ctx context.Context) error {
		external.Infow("starting api",
			"address", conf.API.Address,
		)
		return runServer(ctx, srv)
	})

	// monitor sigint to cancel the context.
	ctx, cancel := junk.WithSignal(ctx, syscall.SIGINT)
	defer cancel()

	// add the database to the queue with it's own launcher and context. we
	// signal it into the other launcher by sending it's result into a task
	// that waits for it.
	db_ctx, db_cancel := context.WithCancel(context.Background())
	defer db_cancel()

	// we subtly do not want a buffer on this channel, because if there were
	// one, we could miss an error from the db.Run call if the cancel call
	// that the manager goroutine defers causes the launcher goroutine to pick
	// that select branch instead of the error channel branch. it's safe to
	// have no buffer because at the end we read from this channel no
	// matter what, which is fine because we close the channel. !!
	db_errch := make(chan error)

	go func() {
		// so that others may know this goroutine exited, and if it has, cancel
		// all the other services
		defer close(db_errch)
		defer cancel()

		db_errch <- junk.Launch(db_ctx, func(ctx context.Context) error {
			external.Infow("starting database")
			return db.Run(ctx)
		})
	}()

	launcher.Queue(func(ctx context.Context) error {
		select {
		case err := <-db_errch:
			return err
		case <-ctx.Done():
		}
		return nil
	})

	// launch everything else
	err = launcher.Run(ctx)

	// ensure that the database is canceled, and that its goroutine is
	// cleaned up.
	db_cancel()
	<-db_errch

	return true, errs.Wrap(err)
}

// runServer runs srv and shuts it down when the context is canceled.
func runServer(ctx context.Context, srv *http.Server) (err error) {
	// I must be going goofy or something, but I see absolutely no way to
	// safely use srv.ListenAndServe() with srv.Shutdown(ctx), because you
	// can't be sure that the ListenAndServe call has started enough for the
	// Shutdown to have an effect. So, let's make a listener ourselves so that
	// we can at least close that on a context cancel, THEN call Shutdown. The
	// bad news about that is, ListenAndServe wraps the listener it makes in
	// a TCP keep alive wrapper, whereas Serve doesn't! That means we get to
	// implement it ourselves. The net/http library is a tire fire.

	lis, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return errs.Wrap(err)
	}
	defer lis.Close()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// monitor when we're done and try to shut down the server.
	go func(ctx context.Context) {
		<-ctx.Done()
		// give the shutdown one minute to clean up
		ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()

		srv.Shutdown(ctx)

		// Just in case we raced and called shutdown before we called Serve.
		// We do this second because we don't want to error in the case that
		// the context is canceled, but if we lose the race, that's ok to
		// error i guess.
		lis.Close()
	}(ctx)

	err = srv.Serve(keepAliveWrapper{lis.(*net.TCPListener)})
	if err == http.ErrServerClosed {
		err = nil
	}
	return errs.Wrap(err)
}

// keepAliveWrapper sets tcp keep alive options on incomming connections.
type keepAliveWrapper struct {
	*net.TCPListener
}

// Listen returns a connection with tcp keep alive options set.
func (k keepAliveWrapper) Accept() (net.Conn, error) {
	conn, err := k.TCPListener.AcceptTCP()
	if err != nil {
		return nil, err
	}
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(3 * time.Minute)
	return conn, nil
}
