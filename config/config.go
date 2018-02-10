// Copyright (C) 2018. See AUTHORS.

package config

import "time"

// Config holds all of the configuration specified inside of a config toml.
type Config struct {
	Main      MainConfig
	Listeners []Entity
	Database  Entity
	API       APIConfig
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
