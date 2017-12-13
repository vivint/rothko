// Copyright (C) 2017. See AUTHORS.

package main

import (
	"io/ioutil"
	"strings"
)

type NameConfig struct {
	Name   string
	Config string
}

type Config struct {
	Plugins     []string     // paths
	Dist        NameConfig   // distribution sketch to use
	Acceptrixes []NameConfig // listeners for packets
	Disk        NameConfig   // disk to handle reads/writes
}

func parseConfig(args []string) (*Config, error) {
	var config Config

	for _, arg := range args {
		parts := strings.SplitN(arg, ":", 3)
		switch parts[0] {
		case "plugin":
			if len(parts) != 2 {
				return nil, InvalidParameters.New(
					`plugin must be of the form "plugin:<path>"`)
			}
			config.Plugins = append(config.Plugins, parts[1])

		case "acceptrix":
			if len(parts) != 3 {
				return nil, InvalidParameters.New(
					`acceptrix must be of the form "acceptrix:<name>:<config>"`)
			}
			config.Acceptrixes = append(config.Acceptrixes, NameConfig{
				Name:   parts[1],
				Config: tryLoadFile(parts[2]),
			})

		case "dist":
			if len(parts) != 3 {
				return nil, InvalidParameters.New(
					`dist must be of the form "dist:<name>:<config>"`)
			}
			config.Dist = NameConfig{
				Name:   parts[1],
				Config: tryLoadFile(parts[2]),
			}

		case "disk":
			if len(parts) != 3 {
				return nil, InvalidParameters.New(
					`disk must be of the form "disk:<name>:<config>"`)
			}
			config.Disk = NameConfig{
				Name:   parts[1],
				Config: tryLoadFile(parts[2]),
			}

		default:
			return nil, InvalidParameters.New(
				"invalid kind: %q",
				parts[0])
		}
	}

	return &config, nil
}

func tryLoadFile(name string) string {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return name
	}
	return string(data)
}
