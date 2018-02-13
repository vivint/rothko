// Copyright (C) 2018. See AUTHORS.

package config

import (
	"os"
	"testing"
	"time"

	"github.com/spacemonkeygo/rothko/internal/assert"
)

func TestLoad(t *testing.T) {
	type D = map[string]interface{}

	conf, err := Load([]byte(InitialConfig))
	assert.NoError(t, err)

	conf.WriteTo(os.Stdout)
	conf.from = nil

	assert.DeepEqual(t, conf, &Config{
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
		Dist: Entity{
			Kind: "tdigest",
			Config: D{
				"compression": float64(5.0),
			},
		},
		API: APIConfig{
			Address: ":8080",
			Domain:  "localhost",
		},
	})
}
