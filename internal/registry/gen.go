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
// Copyright (C) 2017. See AUTHORS.

// created from: go run internal/registry/gen.go %[1]s %[2]s

package %[2]s

import "github.com/spacemonkeygo/rothko/internal/registry"

var (
	// Default is the registry that the exported Register functions use.
	Default Registry

	// Package functions that interact with the Default registry.
	Register = Default.Register
	Lookup   = Default.Lookup
	List     = Default.List
)

// Registration contains information about the registration returned by List.
type Registration = registry.Registration

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

// List returns the set of names that have been registered.
func (r *Registry) List() []Registration {
	return r.reg.List()
}
`

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: go run gen.go <TypeName> <PackageName>")
		os.Exit(1)
	}

	fmt.Printf(strings.TrimSpace(template), os.Args[1], os.Args[2])
	fmt.Println()
}
