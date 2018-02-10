// Copyright (C) 2018. See AUTHORS.

package config

import (
	"github.com/BurntSushi/toml"
)

// Load takes a toml file and parses it into a Config.
func Load(data []byte) (*Config, error) {
	var tomlConfig struct {
		Main struct {
			Duration textDuration
			Plugins  []string
		}
		Listeners map[string][]interface{}
		Database  map[string]interface{}
		API       APIConfig
	}

	if err := toml.Unmarshal(data, &tomlConfig); err != nil {
		return nil, ParseError.Wrap(err)
	}

	if len(tomlConfig.Database) != 1 {
		return nil, ParseError.New("exactly one database must be specified")
	}

	config := &Config{
		Main: MainConfig{
			Duration: tomlConfig.Main.Duration.Duration,
			Plugins:  tomlConfig.Main.Plugins,
		},
		API: tomlConfig.API,
	}

	for kind, data := range tomlConfig.Database {
		config.Database = Entity{
			Kind:   kind,
			Config: data,
		}
		break
	}

	for kind, lis_configs := range tomlConfig.Listeners {
		for _, lis_config := range lis_configs {
			config.Listeners = append(config.Listeners, Entity{
				Kind:   kind,
				Config: lis_config,
			})
		}
	}

	return config, nil
}
