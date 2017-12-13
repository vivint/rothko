// Copyright (C) 2017. See AUTHORS.

//go:generate bash -c "`go list -f '{{ .Dir }}' github.com/spacemonkeygo/rothko`/scripts/regen.sh"
//go:generate bash -c "go run `go list -f '{{ .Dir }}' github.com/spacemonkeygo/rothko/internal/registry`/gen.go DistParamsMaker data > registry.go"

// package data provides a protobuf of the records stored on disk.
package data // import "github.com/spacemonkeygo/rothko/data"
