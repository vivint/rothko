# config
--
    import "github.com/spacemonkeygo/rothko/config"

package config provides methods to load/create the configs.

TODO(jeff): write a bunch of docs around the config toml.

## Usage

```go
var ParseError = errs.Class("parse error")
```
ParseError wraps all of the errors from parsing.

#### type APIConfig

```go
type APIConfig struct {
	Address  string
	Domain   string
	TLS      APITLSConfig
	Security APISecurityConfig
}
```

APIConfig holds configuration for the api config section.

#### type APISecurityConfig

```go
type APISecurityConfig struct {
	Username string
	Password string
}
```

APISecurityConfig holds configuration for the api.security config section.

#### type APITLSConfig

```go
type APITLSConfig struct {
	Key  string
	Cert string
}
```

APITLSConfig holds configuration for the api.tls config section.

#### type Config

```go
type Config struct {
	Main      MainConfig
	Listeners []Entity
	Database  Entity
	Dist      Entity
	API       APIConfig
}
```

Config holds all of the configuration specified inside of a config toml.

#### func  Load

```go
func Load(data []byte) (*Config, error)
```
Load takes a toml file and parses it into a Config.

#### func (*Config) WriteTo

```go
func (c *Config) WriteTo(w io.Writer) error
```

#### type Entity

```go
type Entity struct {
	Kind   string
	Config interface{}
}
```

Entity keeps the kind name as well as the abstract form of the config for
dynamically created entities.

#### type MainConfig

```go
type MainConfig struct {
	Duration time.Duration
	Plugins  []string
}
```

MainConfig holds configuration for the main config section.
