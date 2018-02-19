// Copyright (C) 2018. See AUTHORS.

package config

import (
	"io"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/zeebo/errs"
)

// Config holds all of the configuration specified inside of a config toml.
type Config struct {
	Main      MainConfig
	Listeners []Entity
	Database  Entity
	Dist      Entity
	API       APIConfig

	// keeps track of where the config came from
	from interface{}
}

func (c *Config) WriteTo(w io.Writer) error {
	if c.from == nil {
		return errs.New("config not loaded from Load")
	}
	return errs.Wrap(toml.NewEncoder(w).Encode(c.from))
}

// MainConfig holds configuration for the main config section.
type MainConfig struct {
	Duration time.Duration
	Plugins  []string
}

// Entity keeps the kind name as well as the abstract form of the config
// for dynamically created entities.
type Entity struct {
	Kind   string
	Config interface{}
}

// APIConfig holds configuration for the api config section.
type APIConfig struct {
	Address  string
	Domain   string
	TLS      APITLSConfig
	Security APISecurityConfig
}

// Redact clears out any potentially sensitive data.
func (a APIConfig) Redact() APIConfig {
	a.Security.Username = "redacted"
	a.Security.Password = "redacted"
	a.TLS.Key = "redacted"
	a.TLS.Cert = "redacted"
	return a
}

// APITLSConfig holds configuration for the api.tls config section.
type APITLSConfig struct {
	Key  string
	Cert string
}

// APISecurityConfig holds configuration for the api.security config section.
type APISecurityConfig struct {
	Username string
	Password string
}
