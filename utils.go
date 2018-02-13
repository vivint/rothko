// Copyright (C) 2018. See AUTHORS.

package rothko

import (
	"fmt"
	"strings"

	"github.com/urfave/cli"
)

func t(x string, vs ...interface{}) string {
	return strings.TrimSpace(fmt.Sprintf(x, vs...))
}

func checkArgs(c *cli.Context, expected int) (err error) {
	if c.NArg() != expected {
		err = handled.New("%q requires exactly %d argument(s)",
			c.Command.Name, expected)
		fmt.Printf("Incorrect Usage: %v\n\n", err)
		cli.ShowCommandHelp(c, c.Command.Name)
		return err
	}
	return nil
}
