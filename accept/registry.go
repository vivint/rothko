// Copyright (C) 2017. See AUTHORS.

package accept

import (
	"fmt"
	"sync"
)

var (
	// Default is the registry that the exported Register functions use.
	Default Registry

	// Package functions that interact with the Default registry.
	Register = Default.Register
	Lookup   = Default.Lookup
)

// Registry keeps track of which Acceptrixes are available.
type Registry struct {
	mu  sync.Mutex
	reg map[string]AcceptrixMaker
}

// regLocked returns the registry map, allocating it if necessary.
func (r *Registry) regLocked() map[string]AcceptrixMaker {
	if r.reg == nil {
		r.reg = make(map[string]AcceptrixMaker)
	}
	return r.reg
}

// Register registers the AcceptrixMaker under the given name. It panics if
// it already has one for the given name.
func (r *Registry) Register(name string, maker AcceptrixMaker) {
	r.mu.Lock()
	defer r.mu.Unlock()

	reg := r.regLocked()
	if _, ok := reg[name]; ok {
		panic(fmt.Sprintf("%q already registered", name))
	}
	reg[name] = maker
}

// Lookup returns the AcceptrixMaker for the given name if one exists and nil
// if one does not exist.
func (r *Registry) Lookup(name string) AcceptrixMaker {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.regLocked()[name]
}
