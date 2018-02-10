// Copyright (C) 2018. See AUTHORS.

package config

import (
	"testing"
	"time"

	"github.com/spacemonkeygo/rothko/internal/assert"
)

func TestLoad(t *testing.T) {
	type D = map[string]interface{}

	config, err := Load([]byte(initialConfig))
	assert.NoError(t, err)
	assert.DeepEqual(t, config, &Config{
		Main: MainConfig{
			Duration: 10 * time.Minute,
			Plugins:  []string{},
		},
		Listeners: []Entity{{
			Kind: "graphite",
			Config: D{
				"address": ":1111",
			},
		}},
		Database: Entity{
			Kind: "files",
			Config: D{
				"directory": "data",
				"size":      int64(256),
				"cap":       int64(400),
				"files":     int64(2),
			},
		},
		API: APIConfig{
			Address: ":8080",
			Domain:  "localhost",
		},
	})
}
