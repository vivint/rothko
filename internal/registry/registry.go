// Copyright (C) 2017. See AUTHORS.

package registry

import (
	"fmt"
	"sync"
)

// Registry keeps track of things.
type Registry struct {
	mu  sync.Mutex
	reg map[string]interface{}
}

// regLocked returns the registry map, allocating it if necessary.
func (r *Registry) regLocked() map[string]interface{} {
	if r.reg == nil {
		r.reg = make(map[string]interface{})
	}
	return r.reg
}

// Register registers the thing under the given name. It panics if
// it already has one for the given name.
func (r *Registry) Register(name string, thing interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	reg := r.regLocked()
	if _, ok := reg[name]; ok {
		panic(fmt.Sprintf("%q already registered", name))
	}
	reg[name] = thing
}

// Lookup returns the thing for the given name if one exists and nil
// if one does not exist.
func (r *Registry) Lookup(name string) interface{} {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.regLocked()[name]
}
