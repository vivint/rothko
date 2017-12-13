// Copyright (C) 2017. See AUTHORS.

package registry

import (
	"fmt"
	"runtime"
	"sort"
	"sync"
)

// Registration contains information about the registration returned by List.
type Registration struct {
	Name      string
	Registrar string
}

// stored in the registry
type thing struct {
	value     interface{}
	registrar string
}

// Registry keeps track of things.
type Registry struct {
	mu  sync.Mutex
	reg map[string]thing
}

// regLocked returns the registry map, allocating it if necessary.
func (r *Registry) regLocked() map[string]thing {
	if r.reg == nil {
		r.reg = make(map[string]thing)
	}
	return r.reg
}

// Register registers the value under the given name. It panics if
// it already has one for the given name.
func (r *Registry) Register(name string, value interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	reg := r.regLocked()
	if _, ok := reg[name]; ok {
		panic(fmt.Sprintf("%q already registered", name))
	}

	registrar := "<unknown>"
	if pc, _, line, ok := runtime.Caller(3); ok {
		frame, _ := runtime.CallersFrames([]uintptr{pc}).Next()
		if frame.Func != nil {
			registrar = fmt.Sprintf("%s:%d", frame.Func.Name(), line)
		}
	}

	reg[name] = thing{
		value:     value,
		registrar: registrar,
	}
}

// Lookup returns the value for the given name if one exists and nil
// if one does not exist.
func (r *Registry) Lookup(name string) interface{} {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.regLocked()[name].value
}

// List returns the set of names that have been registered.
func (r *Registry) List() []Registration {
	r.mu.Lock()
	defer r.mu.Unlock()

	reg := r.regLocked()

	regs := make([]Registration, 0, len(reg))
	for name, thing := range reg {
		regs = append(regs, Registration{
			Name:      name,
			Registrar: thing.registrar,
		})
	}
	sort.Slice(regs, func(i, j int) bool {
		return regs[i].Name < regs[j].Name
	})

	return regs
}
