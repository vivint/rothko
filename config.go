// Copyright (C) 2017. See AUTHORS.

package rothko

import (
	"context"
	"io/ioutil"
	"plugin"
	"strings"

	"github.com/spacemonkeygo/rothko/accept"
	"github.com/spacemonkeygo/rothko/data"
	"github.com/spacemonkeygo/rothko/data/scribble"
	"github.com/spacemonkeygo/rothko/disk"
	"github.com/zeebo/errs"
)

// NameConfig is a tuple of a Name and a Config.
type NameConfig struct {
	Name   string
	Config string
}

// Config describes the resources used during operation.
type Config struct {
	Plugins     []string     // paths
	Dist        NameConfig   // distribution sketch to use
	Acceptrixes []NameConfig // listeners for packets
	Disk        NameConfig   // disk to handle reads/writes
}

// ParseConfig takes the arguments and parses it into a Config.
func ParseConfig(args []string) (*Config, []string, error) {
	var conf Config

loop:
	for ; len(args) > 0; args = args[1:] {
		arg := args[0]

		parts := strings.SplitN(arg, ":", 3)
		switch parts[0] {
		case "plugin":
			if len(parts) != 2 {
				return nil, nil, ErrInvalidParameters.New(
					`plugin must be of the form "plugin:<path>"`)
			}
			conf.Plugins = append(conf.Plugins, parts[1])

		case "acceptrix":
			if len(parts) != 3 {
				return nil, nil, ErrInvalidParameters.New(
					`acceptrix must be of the form "acceptrix:<name>:<config>"`)
			}
			conf.Acceptrixes = append(conf.Acceptrixes, NameConfig{
				Name:   parts[1],
				Config: tryLoadFile(parts[2]),
			})

		case "dist":
			if len(parts) != 3 {
				return nil, nil, ErrInvalidParameters.New(
					`dist must be of the form "dist:<name>:<config>"`)
			}
			conf.Dist = NameConfig{
				Name:   parts[1],
				Config: tryLoadFile(parts[2]),
			}

		case "disk":
			if len(parts) != 3 {
				return nil, nil, ErrInvalidParameters.New(
					`disk must be of the form "disk:<name>:<config>"`)
			}
			conf.Disk = NameConfig{
				Name:   parts[1],
				Config: tryLoadFile(parts[2]),
			}

		default:
			break loop
		}
	}

	return &conf, args, nil
}

func tryLoadFile(name string) string {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return name
	}
	return string(data)
}

// LoadPlugins loads all of the plugins specified by the config.
func (c *Config) LoadPlugins() error {
	for _, path := range c.Plugins {
		if _, err := plugin.Open(path); err != nil {
			return errs.Wrap(err)
		}
	}
	return nil
}

// LoadDisk loads the Disk from the config.
func (c *Config) LoadDisk(ctx context.Context) (disk.Disk, error) {
	disk_maker := disk.Lookup(c.Disk.Name)
	if disk_maker == nil {
		return nil, ErrMissing.New("unknown disk: %q", c.Dist.Name)
	}
	di, err := disk_maker(ctx, c.Disk.Config)
	if err != nil {
		return nil, errs.Wrap(err)
	}
	return di, nil
}

// LoadAcceptrixes loads the acceptrixes from the config.
func (c *Config) LoadAcceptrixes(ctx context.Context) (
	accs []accept.Acceptrix, err error) {

	for _, nc := range c.Acceptrixes {
		acc_maker := accept.Lookup(nc.Name)
		if acc_maker == nil {
			return nil, ErrMissing.New("unknown acceptrix: %q", nc.Name)
		}
		acc, err := acc_maker(ctx, nc.Config)
		if err != nil {
			return nil, errs.Wrap(err)
		}
		accs = append(accs, acc)
	}

	return accs, nil
}

// LoadScribbler loads the scribbler from the config.
func (c *Config) LoadScribbler(ctx context.Context) (
	*scribble.Scribbler, error) {

	dist_maker := data.Lookup(c.Dist.Name)
	if dist_maker == nil {
		return nil, ErrMissing.New("unknown dist: %q", c.Dist.Name)
	}
	params, err := dist_maker(ctx, c.Dist.Config)
	if err != nil {
		return nil, errs.Wrap(err)
	}
	return scribble.NewScribbler(params), nil
}
