// Copyright (C) 2017. See AUTHORS.

// created from: go run internal/registry/gen.go DiskMaker disk

package disk

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

// Registry keeps track of DiskMaker values by their name.
type Registry struct {
	reg registry.Registry
}

// Register adds a DiskMaker value to the registry under the name, and
// panics if the name already exists.
func (r *Registry) Register(name string, value DiskMaker) {
	r.reg.Register(name, value)
}

// Lookup returns the DiskMaker for the name, or the zero value if
// nothing exists.
func (r *Registry) Lookup(name string) DiskMaker {
	out, _ := r.reg.Lookup(name).(DiskMaker)
	return out
}

// List returns the set of names that have been registered.
func (r *Registry) List() []Registration {
	return r.reg.List()
}
