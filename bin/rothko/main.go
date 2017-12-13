// Copyright (C) 2017. See AUTHORS.

// rothko runs a rothko server with standard rothko implementations.
package main // import "github.com/spacemonkeygo/rothko/bin/rothko"

import (
	"github.com/spacemonkeygo/rothko"
	_ "github.com/spacemonkeygo/rothko/data/dists/tdigest"
	_ "github.com/spacemonkeygo/rothko/disk/files"
)

func main() { rothko.Main() }
