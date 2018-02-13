// Copyright (C) 2018. See AUTHORS.

package registry

import (
	"context"

	"github.com/spacemonkeygo/rothko/database"
	"github.com/spacemonkeygo/rothko/dist"
	"github.com/spacemonkeygo/rothko/listener"
)

// Default is the default registry that the Register calls insert into.
var Default Registry

// RegisterListener registers the ListenerMaker as the provided name in
// the Default registry. It overwrites any previous calls for the same name.
func RegisterListener(name string, maker ListenerMaker) {
	Default.RegisterListener(name, maker)
}

// NewListener constructs a Listener using the ListenerMaker registered under
// the name in the Default registry.
// It returns an error if there has been no such registration.
func NewListener(ctx context.Context, name string, config interface{}) (
	listener.Listener, error) {
	return Default.NewListener(ctx, name, config)
}

// RegisterDatabase registers the DatabaseMaker as the provided name in
// the Default registry. It overwrites any previous calls for the same name.
func RegisterDatabase(name string, maker DatabaseMaker) {
	Default.RegisterDatabase(name, maker)
}

// NewDatabase constructs a Database using the DatabaseMaker registered under
// the name in the Default registry.
// It returns an error if there has been no such registration.
func NewDatabase(ctx context.Context, name string, config interface{}) (
	database.DB, error) {
	return Default.NewDatabase(ctx, name, config)
}

// RegisterDistribution registers the DistributionMaker as the provided name in
// the Default registry. It overwrites any previous calls for the same name.
func RegisterDistribution(name string, maker DistributionMaker) {
	Default.RegisterDistribution(name, maker)
}

// NewDistribution constructs a Distribution using the DistributionMaker
// registered under the name in the Default registry.
// It returns an error if there has been no such registration.
func NewDistribution(ctx context.Context, name string, config interface{}) (
	dist.Params, error) {
	return Default.NewDistribution(ctx, name, config)
}

//
// Listeners
//

// ListenerMaker constructs a listener from the provided config.
type ListenerMaker interface {
	New(ctx context.Context, config interface{}) (listener.Listener, error)
}

// ListenerMakerFunc is a function type that implements ListenerMaker.
type ListenerMakerFunc func(context.Context, interface{}) (
	listener.Listener, error)

// New calls the ListenerMakerFunc.
func (fn ListenerMakerFunc) New(ctx context.Context, config interface{}) (
	listener.Listener, error) {

	return fn(ctx, config)
}

//
// Databases
//

// DatabaseMaker constructs a listener from the provided config.
type DatabaseMaker interface {
	New(ctx context.Context, config interface{}) (database.DB, error)
}

// DatabaseMakerFunc is a function type that implements DatabaseMaker.
type DatabaseMakerFunc func(context.Context, interface{}) (database.DB, error)

// New calls the DatabaseMakerFunc.
func (fn DatabaseMakerFunc) New(ctx context.Context, config interface{}) (
	database.DB, error) {

	return fn(ctx, config)
}

//
// Distributions
//

// DistributionMaker constructs a listener from the provided config.
type DistributionMaker interface {
	New(ctx context.Context, config interface{}) (dist.Params, error)
}

// DistributionMakerFunc is a function type that implements DistributionMaker.
type DistributionMakerFunc func(context.Context, interface{}) (
	dist.Params, error)

// New calls the DistributionMakerFunc.
func (fn DistributionMakerFunc) New(ctx context.Context, config interface{}) (
	dist.Params, error) {

	return fn(ctx, config)
}
