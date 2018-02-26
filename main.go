// Copyright (C) 2018. See AUTHORS.

package rothko

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
	_ "github.com/vivint/rothko/database/files"
	_ "github.com/vivint/rothko/dist/tdigest"
	_ "github.com/vivint/rothko/listener/graphite"
	"github.com/zeebo/errs"
)

var handled = errs.Class("")

// Main is the entrypoint to any rothko binary. It is exposed so that it is
// easy to create custom binaries with your own enhancements.
func Main() {
	app := cli.NewApp()
	app.Usage = "a time-distribution metric store"
	app.Version = "0.0.1"

	app.Commands = []cli.Command{
		initCommand,
		runCommand,
	}

	if err := app.Run(os.Args); err != nil {
		if !handled.Has(err) {
			fmt.Printf("unexpected error: %+v\n", err)
		}
		os.Exit(1)
	}
	os.Exit(0)
}
