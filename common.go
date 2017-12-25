// Copyright (C) 2017. See AUTHORS.

package rothko

import (
	"fmt"
	"io"
	"strings"

	"github.com/spacemonkeygo/rothko/accept"
	"github.com/spacemonkeygo/rothko/data"
	"github.com/spacemonkeygo/rothko/disk"
	"github.com/spacemonkeygo/rothko/internal/junk"
	"github.com/zeebo/errs"
)

// error classes that are used in certain circumstances
var (
	ErrInvalidParameters = errs.Class("invalid parameters")
	ErrMissing           = errs.Class("missing")
)

func printUsage(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
usage: rothko [list|help] [parameters...]

parameters are of the form <kind>:<value> and there are four kinds:

	plugin:    pass "plugin:<path>" to load the plugin
	acceptrix: pass "acceptrix:<name>:<config>" to add an acceptrix
	disk:      pass "disk:<name>:<config>" to use the disk
	dist:      pass "dist:<name>:<config>" to use the distribution sketch

disk and dist are required. config may either be a string literal or a path to
a file containing the data.

the acceptrix is used to read data typically from a network interface and add
it to the disk using the distribution sketch. there may be multiple acceptrix
declarations.

for example:

	rothko \
		plugin:spacemonkey.so \
		acceptrix:sm/collector:0.0.0.0:9000 \
		disk:rothko/disk/files:files.json \
		dist:rothko/dist/tdigest:compression=5

will load the spacemonkey.so plugin, use the sm/collector acceptrix instructed
to listen on 0.0.0.0:9000, use the rothko files database configured from
files.json, and use the tdigest sketch with a compression of 5.

if you run "rothko list" and pass a set of plugins, the set of registered
acceptrixes, dists, and disks are outputted. run "rothko help" to see this
message.
`))
}

// listAvailable is a tiny helper to print a tab aligned list of available
// entities that can be used.
func listAvailable(w io.Writer) {
	tw := junk.NewTabbed(w)
	tw.Write("kind", "name", "registrar")
	for _, reg := range accept.List() {
		tw.Write("acceptrix", reg.Name, reg.Registrar)
	}
	for _, reg := range disk.List() {
		tw.Write("disk", reg.Name, reg.Registrar)
	}
	for _, reg := range data.List() {
		tw.Write("dist", reg.Name, reg.Registrar)
	}
	tw.Flush()
}
