// Copyright (C) 2017. See AUTHORS.

package rothko

import (
	"io/ioutil"
	"strings"
)

type nameConfig struct {
	Name   string
	Config string
}

type config struct {
	Plugins     []string     // paths
	Dist        nameConfig   // distribution sketch to use
	Acceptrixes []nameConfig // listeners for packets
	Disk        nameConfig   // disk to handle reads/writes
}

func parseConfig(args []string) (*config, error) {
	var conf config

	for _, arg := range args {
		parts := strings.SplitN(arg, ":", 3)
		switch parts[0] {
		case "plugin":
			if len(parts) != 2 {
				return nil, errInvalidParameters.New(
					`plugin must be of the form "plugin:<path>"`)
			}
			conf.Plugins = append(conf.Plugins, parts[1])

		case "acceptrix":
			if len(parts) != 3 {
				return nil, errInvalidParameters.New(
					`acceptrix must be of the form "acceptrix:<name>:<config>"`)
			}
			conf.Acceptrixes = append(conf.Acceptrixes, nameConfig{
				Name:   parts[1],
				Config: tryLoadFile(parts[2]),
			})

		case "dist":
			if len(parts) != 3 {
				return nil, errInvalidParameters.New(
					`dist must be of the form "dist:<name>:<config>"`)
			}
			conf.Dist = nameConfig{
				Name:   parts[1],
				Config: tryLoadFile(parts[2]),
			}

		case "disk":
			if len(parts) != 3 {
				return nil, errInvalidParameters.New(
					`disk must be of the form "disk:<name>:<config>"`)
			}
			conf.Disk = nameConfig{
				Name:   parts[1],
				Config: tryLoadFile(parts[2]),
			}

		default:
			return nil, errInvalidParameters.New(
				"invalid kind: %q",
				parts[0])
		}
	}

	return &conf, nil
}

func tryLoadFile(name string) string {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return name
	}
	return string(data)
}
