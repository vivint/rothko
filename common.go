// Copyright (C) 2018. See AUTHORS.

package rothko

import (
	"context"

	"github.com/spacemonkeygo/rothko/data"
	"github.com/spacemonkeygo/rothko/data/scribble"
	"github.com/spacemonkeygo/rothko/disk"
)

// Acceptrix is a type that reads from some data source and pushes the data
// into the scribbler.
type Acceptrix interface {
	// Run should scribble the data into the provided Scribbler until the
	// context is canceled.
	Run(ctx context.Context, scr *scribble.Scribbler) error
}

// Dumper periodically dumps the scribbler.
type Dumper interface {
	Run(ctx context.Context, scr *scribble.Scribbler) error
}

// Option is a way to specify a set of options.
type Option func(*Options)

type Options struct {
	Dumper      Dumper
	Disk        disk.Disk
	DistParams  data.DistParams
	Acceptrixes []Acceptrix
}

func WithDumper(dumper Dumper) Option {
	return func(o *Options) { o.Dumper = dumper }
}

func WithDisk(disk disk.Disk) Option {
	return func(o *Options) { o.Disk = disk }
}

func WithAcceptrixes(acceptrixes ...Acceptrix) Option {
	return func(o *Options) { o.Acceptrixes = acceptrixes }
}

func WithDistParams(params data.DistParams) Option {
	return func(o *Options) { o.DistParams = params }
}
