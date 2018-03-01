// Copyright (C) 2018. See AUTHORS.

package rothko

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"plugin"
	"syscall"
	"time"

	"github.com/urfave/cli"
	"github.com/vivint/rothko/api"
	"github.com/vivint/rothko/config"
	"github.com/vivint/rothko/data"
	"github.com/vivint/rothko/dump"
	"github.com/vivint/rothko/external"
	"github.com/vivint/rothko/internal/junk"
	"github.com/vivint/rothko/internal/tgzfs"
	"github.com/vivint/rothko/internal/tmplfs"
	"github.com/vivint/rothko/registry"
	"github.com/vivint/rothko/ui"
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

// run creates and starts all of the services defined by the config. It exits
// when the context is canceled, or when an appropriate signal is sent to the
// binary. The started return value is true if the services were created and
// started before returning.
func run(ctx context.Context, conf *config.Config) (started bool, err error) {
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

	// create the writer
	w := data.NewWriter(params)

	// create the dumper
	dumper := dump.New(dump.Options{
		DB:     db,
		Period: conf.Main.Duration,
	})

	// create the api server
	// TODO(jeff): basic auth
	// TODO(jeff): tls
	// TODO(jeff): proper CORS
	external.Infow("creating api",
		"config", conf.API.Redact(),
	)

	var static http.Handler = tarballWarning{}
	if ui.Tarball != nil {
		fs, err := tgzfs.New(ui.Tarball)
		if err != nil {
			return false, errs.Wrap(err)
		}
		static = tmplfs.New(fs)
	}
	srv := &http.Server{
		Addr:    conf.API.Address,
		Handler: api.New(db, static),
	}

	// create and queue the listeners
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

	// queue the worker that periodically dumps in to the database
	launcher.Queue(func(ctx context.Context) error {
		external.Infow("starting dumper")
		return dumper.Run(ctx, w)
	})

	// queue the api server
	launcher.Queue(func(ctx context.Context) error {
		external.Infow("starting api",
			"config", conf.API.Redact(),
		)
		return runServer(ctx, srv)
	})

	// because we don't want to rerun the database on sigint, we launch all of
	// the services under one launcher with the sigint_ctx, and launch the
	// database under another launcher with a parent of the sigint_ctx.
	//
	// in the case the sigint_ctx launcher returns no error, we run a dump with
	// a timeout of 60 seconds, before returning and causing all of the other
	// tasks in the launcher to be canceled.

	var parent junk.Launcher

	// queue all the other services
	parent.Queue(func(ctx context.Context) error {
		sigint_ctx, cancel := junk.WithSignal(ctx, syscall.SIGINT)
		defer cancel()

		if err := launcher.Run(sigint_ctx); err != nil {
			return err
		}

		// run with the parent of the sigint_ctx so that we get canceled if
		// the database Run exits.
		ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
		defer cancel()

		external.Infow("performing last dump")
		dumper.Dump(ctx, w)
		return nil
	})

	// queue the database
	parent.Queue(func(ctx context.Context) error {
		external.Infow("starting database",
			"kind", conf.Database.Kind,
			"config", conf.Database.Config,
		)
		return db.Run(ctx)
	})

	// run our stuff
	return true, errs.Wrap(parent.Run(ctx))
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
		ctx, cancel := context.WithTimeout(
			context.Background(), 60*time.Second)
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

// tarballWarning is an http.Handler that is served for the static site if a
// ui tarball has not been generated.
type tarballWarning struct{}

// ServeHTTP implements the http.Handler interface, responding with a warning
// if index.html is requested.
func (tarballWarning) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/index.html", "index.html", "/", "":
	default:
		http.NotFound(w, req)
		return
	}

	io.WriteString(w, `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8"/>
	<title>Rothko</title>
</head>
<body>
	<h1>Rothko</h1>
	<p>You have not generated the ui in this build of Rothko.</p>
	<p>If you would like to have a nice ui, run
	   <span style="font-family: monospace">go generate</span> on the
	   <span style="font-family: monospace">github.com/vivint/rothko/ui</span>
	   package, and re-build your binary.
	</p>
</body>
</html>
`)
}
