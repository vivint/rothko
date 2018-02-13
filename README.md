# package rothko

`import "github.com/spacemonkeygo/rothko"`

package rothko is a time-distribution system.

It stores and allows interaction with distributions of a metric that vary
through time. This allows you to collect insight about the overall values of
metrics when there are many values from multiple hosts.

This package contains a Main function to be called by the actual main package,
allowing you to customize the configuration of rothko's operation.

See the github.com/spacemonkeygo/rothko/bin/rothko package for a rothko binary
with implementations loaded from this project.

## Usage

#### func  Main

```go
func Main(conf config.Config)
```
Main is the entrypoint to any rothko binary. It is exposed so that it is easy to
create custom binaries with your own enhancements.
