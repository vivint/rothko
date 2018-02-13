// Copyright (C) 2018. See AUTHORS.

package rothko

import (
	"fmt"
	"os"

	_ "github.com/spacemonkeygo/rothko/database/files"
	_ "github.com/spacemonkeygo/rothko/dist/tdigest"
	_ "github.com/spacemonkeygo/rothko/listener/graphite"
	"github.com/urfave/cli"
	"github.com/zeebo/errs"
)

var handled = errs.Class("")

// Main is the entrypoint to any rothko binary. It is exposed so that it is
// easy to create custom binaries with your own enhancements.
func Main() {
	app := cli.NewApp()
	app.Usage = "a time-database metric store"
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
}
