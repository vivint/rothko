// Copyright (C) 2017. See AUTHORS.
//
// gen is used to build a concrete registry for a given type.
//
// usage: go run gen.go <TypeName>

// +build ignore

package main

import (
	"fmt"
	"os"
	"strings"
)

const template = `
import "github.com/spacemonkeygo/rothko/internal/registry"

var (
	// Default is the registry that the exported Register functions use.
	Default Registry

	// Package functions that interact with the Default registry.
	Register = Default.Register
	Lookup   = Default.Lookup
)

// Registry keeps track of %[1]s values by their name.
type Registry struct {
	reg registry.Registry
}

// Register adds a %[1]s value to the registry under the name, and
// panics if the name already exists.
func (r *Registry) Register(name string, value %[1]s) {
	r.reg.Register(name, value)
}

// Lookup returns the %[1]s for the name, or the zero value if
// nothing exists.
func (r *Registry) Lookup(name string) %[1]s {
	out, _ := r.reg.Lookup(name).(%[1]s)
	return out
}
`

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go run gen.go <TypeName>")
		os.Exit(1)
	}

	fmt.Printf(strings.TrimSpace(template), os.Args[1])
	fmt.Println()
}
