// Copyright (C) 2018. See AUTHORS.

package registry

import (
	"context"
	"sync"

	"github.com/vivint/rothko/database"
	"github.com/vivint/rothko/dist"
	"github.com/vivint/rothko/listener"
	"github.com/zeebo/errs"
)

// Registry keeps track of a set of Makers by name.
type Registry struct {
	mu sync.Mutex

	listeners     map[string]ListenerMaker
	databases     map[string]DatabaseMaker
	distributions map[string]DistributionMaker
}

// RegisterListener registers the ListenerMaker under the given name.
// It overwrites any previous calls for the same name.
func (r *Registry) RegisterListener(name string, maker ListenerMaker) {
	r.mu.Lock()
	if r.listeners == nil {
		r.listeners = make(map[string]ListenerMaker)
	}
	r.listeners[name] = maker
	r.mu.Unlock()
}

// NewListener constructs a Listener using the ListenerMaker registered under
// the name. It returns an error if there has been no such registration.
func (r *Registry) NewListener(ctx context.Context, name string,
	config interface{}) (listener.Listener, error) {

	r.mu.Lock()
	defer r.mu.Unlock()

	maker, ok := r.listeners[name]
	if !ok {
		return nil, errs.New("no registration for: %q", name)
	}

	return maker.New(ctx, config)
}

// RegisterDatabase registers the DatabaseMaker under the given name.
// It overwrites any previous calls for the same name.
func (r *Registry) RegisterDatabase(name string, maker DatabaseMaker) {
	r.mu.Lock()
	if r.databases == nil {
		r.databases = make(map[string]DatabaseMaker)
	}
	r.databases[name] = maker
	r.mu.Unlock()
}

// NewDatabase constructs a Database using the DatabaseMaker registered under
// the name. It returns an error if there has been no such registration.
func (r *Registry) NewDatabase(ctx context.Context, name string,
	config interface{}) (database.DB, error) {

	r.mu.Lock()
	defer r.mu.Unlock()

	maker, ok := r.databases[name]
	if !ok {
		return nil, errs.New("no registration for: %q", name)
	}

	return maker.New(ctx, config)
}

// RegisterDistribution registers the DistributionMaker under the given name.
// It overwrites any previous calls for the same name.
func (r *Registry) RegisterDistribution(name string, maker DistributionMaker) {
	r.mu.Lock()
	if r.distributions == nil {
		r.distributions = make(map[string]DistributionMaker)
	}
	r.distributions[name] = maker
	r.mu.Unlock()
}

// NewDistribution constructs a Distribution using the DistributionMaker
// registered under the name. It returns an error if there has been no such
// registration.
func (r *Registry) NewDistribution(ctx context.Context, name string,
	config interface{}) (dist.Params, error) {

	r.mu.Lock()
	defer r.mu.Unlock()

	maker, ok := r.distributions[name]
	if !ok {
		return nil, errs.New("no registration for: %q", name)
	}

	return maker.New(ctx, config)
}
