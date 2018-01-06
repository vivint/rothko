// Copyright (C) 2017. See AUTHORS.

// rothko runs a demo rothko server with standard rothko implementations.
package main // import "github.com/spacemonkeygo/rothko/bin/rothko"

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/spacemonkeygo/rothko"
	"github.com/spacemonkeygo/rothko/data/dists/tdigest"
	"github.com/spacemonkeygo/rothko/disk/files"
	"github.com/spacemonkeygo/rothko/dump"
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
	dumper := dump.New(dump.Options{
		Disk:   disk,
		Period: 10 * time.Minute,
	})

	fmt.Println(`
This demonstration binary fills the database with bogus random data which can
be viewed at <TODO: make a web ui :)>. In a real production deployment, a
way to ingest data and write it would be required. See the rothko.Acceptrix
interface for how.`)

	rothko.Main(
		rothko.WithDisk(disk),
		rothko.WithDistParams(params),
		rothko.WithDumper(dumper))
}
