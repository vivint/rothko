// Copyright (C) 2017. See AUTHORS.

// rothko runs a rothko server with standard rothko implementations.
package main // import "github.com/spacemonkeygo/rothko/bin/rothko"

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spacemonkeygo/rothko"
	"github.com/spacemonkeygo/rothko/data/dists/tdigest"
	"github.com/spacemonkeygo/rothko/disk/files"
)

var (
	filesConfigPath = flag.String("files_config", "files.json",
		"path to json file containing the config for the files backend")

	compression = flag.Float64("compression", 5,
		"compression value to use for t-digest")
)

func main() {
	flag.Parse()

	files_config_data, err := ioutil.ReadFile(*filesConfigPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "unable to open files_config:", err)
		os.Exit(1)
	}

	var filesOptions struct {
		files.Options
		Dir string
	}

	if err := json.Unmarshal(files_config_data, &filesOptions); err != nil {
		fmt.Fprintln(os.Stderr, "unable to load json for files backend:", err)
		os.Exit(1)
	}

	disk := files.New(filesOptions.Dir, filesOptions.Options)
	params := tdigest.Params{Compression: *compression}

	rothko.Main(
		rothko.WithDisk(disk),
		rothko.WithDistParams(params))
}
