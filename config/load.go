// Copyright (C) 2018. See AUTHORS.

package config

import (
	"github.com/BurntSushi/toml"
)

// Load takes a toml file and parses it into a Config.
func Load(data []byte) (*Config, error) {
	// tomlConfig somewhat mirrors Config except it contains tags to marshal
	// the values in lower case, as well as using some dynamic stuff for
	// entities that can be added by plugins.
	var tomlConfig struct {
		Main struct {
			Duration textDuration `toml:"duration"`
			Plugins  []string     `toml:"plugins"`
		} `toml:"main"`
		Listeners map[string][]interface{} `toml:"listeners"`
		Database  map[string]interface{}   `toml:"database"`
		Dist      map[string]interface{}   `toml:"dist"`
		API       struct {
			Address string `toml:"address"`
			Origin  string `toml:"origin"`
			TLS     struct {
				Key  string `toml:"key"`
				Cert string `toml:"cert"`
			} `toml:"tls"`
			Security struct {
				Username string `toml:"username"`
				Password string `toml:"password"`
			} `toml:"security"`
		} `toml:"api"`
	}

	if err := toml.Unmarshal(data, &tomlConfig); err != nil {
		return nil, ParseError.Wrap(err)
	}

	if len(tomlConfig.Database) != 1 {
		return nil, ParseError.New("exactly one database must be specified")
	}

	if len(tomlConfig.Dist) != 1 {
		return nil, ParseError.New("exactly one dist must be specified")
	}

	conf := &Config{
		from: tomlConfig,

		Main: MainConfig{
			Duration: tomlConfig.Main.Duration.Duration,
			Plugins:  tomlConfig.Main.Plugins,
		},
		API: APIConfig{
			Address:  tomlConfig.API.Address,
			Origin:   tomlConfig.API.Origin,
			TLS:      APITLSConfig(tomlConfig.API.TLS),
			Security: APISecurityConfig(tomlConfig.API.Security),
		},
	}

	for kind, config := range tomlConfig.Database {
		conf.Database = Entity{
			Kind:   kind,
			Config: config,
		}
		break
	}

	for kind, config := range tomlConfig.Dist {
		conf.Dist = Entity{
			Kind:   kind,
			Config: config,
		}
		break
	}

	for kind, configs := range tomlConfig.Listeners {
		for _, config := range configs {
			conf.Listeners = append(conf.Listeners, Entity{
				Kind:   kind,
				Config: config,
			})
		}
	}

	return conf, nil
}
