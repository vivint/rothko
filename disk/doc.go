// Copyright (C) 2017. See AUTHORS.

//go:generate bash -c "go run `go list -f '{{ .Dir }}' github.com/spacemonkeygo/rothko/internal/registry`/gen.go DiskMaker disk > registry.go"

// package disk provides interfaces to disk storage of records.
package disk // import "github.com/spacemonkeygo/rothko/disk"
