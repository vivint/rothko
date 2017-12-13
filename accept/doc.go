// Copyright (C) 2017. See AUTHORS.

//go:generate bash -c "go run `go list -f '{{ .Dir }}' github.com/spacemonkeygo/rothko/internal/registry`/gen.go AcceptrixMaker accept > registry.go"

// package accept provides interfaces and a registry to accept data
package accept // import "github.com/spacemonkeygo/rothko/accept"
