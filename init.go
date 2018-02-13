// Copyright (C) 2018. See AUTHORS.

package rothko

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spacemonkeygo/rothko/config"
	"github.com/urfave/cli"
	"github.com/zeebo/errs"
)

const configPath = "rothko.toml"

var initCommand = cli.Command{
	Name:  "init",
	Usage: "create a new configuration",
	ArgsUsage: t(`
`),

	Description: t(`
The init command will create a new file named %q. It is meant to be
edited, but contains useful defaults.
`, configPath),

	Action: func(c *cli.Context) error {
		if err := checkArgs(c, 0); err != nil {
			return err
		}

		_, err := os.Stat(configPath)
		switch {
		case os.IsNotExist(err):
		case err == nil:
			fmt.Printf("config file already exists. remove %q first\n",
				configPath)
			return handled.New("")
		case err != nil:
			return errs.Wrap(err)
		}

		err = ioutil.WriteFile(
			configPath, []byte(config.InitialConfig), 0644)
		if err != nil {
			return errs.Wrap(err)
		}

		fmt.Printf("wrote initial config to %q\n", configPath)
		return nil
	},
}
