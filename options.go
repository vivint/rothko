// Copyright (C) 2017. See AUTHORS.

package rothko

import (
	"github.com/spacemonkeygo/rothko/data"
	"github.com/spacemonkeygo/rothko/disk"
)

// Option is a way to specify a set of options.
type Option func(*Options)

type Options struct {
	Logger      Logger
	Config      Config
	Monitor     Monitor
	Disk        disk.Disk
	DistParams  data.DistParams
	Acceptrixes []Acceptrix
}

func WithLogger(logger Logger) Option {
	return func(o *Options) { o.Logger = logger }
}

func WithConfig(config Config) Option {
	return func(o *Options) { o.Config = config }
}

func WithMonitor(monitor Monitor) Option {
	return func(o *Options) { o.Monitor = monitor }
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
