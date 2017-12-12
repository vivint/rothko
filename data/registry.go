// Copyright (C) 2017. See AUTHORS.

package data

import "github.com/spacemonkeygo/rothko/internal/registry"

var (
	// Default is the registry that the exported Register functions use.
	Default Registry

	// Package functions that interact with the Default registry.
	Register = Default.Register
	Lookup   = Default.Lookup
)

// Registry keeps track of DistMaker values by their name.
type Registry struct {
	reg registry.Registry
}

// Register adds a DistMaker value to the registry under the name, and
// panics if the name already exists.
func (r *Registry) Register(name string, value DistMaker) {
	r.reg.Register(name, value)
}

// Lookup returns the DistMaker for the name, or the zero value if
// nothing exists.
func (r *Registry) Lookup(name string) DistMaker {
	out, _ := r.reg.Lookup(name).(DistMaker)
	return out
}
