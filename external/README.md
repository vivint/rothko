# external
--
    import "github.com/spacemonkeygo/rothko/external"

package external defines some interfaces for external resources.

## Usage

#### func  Errorw

```go
func Errorw(msg string, keyvals ...interface{})
```
Errorw calls Errorw on the default Resources.

#### func  Infow

```go
func Infow(msg string, keyvals ...interface{})
```
Infow calls Infow on the default resources.

#### func  Observe

```go
func Observe(name string, value float64)
```
Observe calls Observe on the default Resources.

#### type Logger

```go
type Logger interface {
	Infow(msg string, keyvals ...interface{})
	Errorw(msg string, keyvals ...interface{})
}
```

Logger is used when logging is required. It is built to match the uber/zap
SugaredLogger type.

#### type Monitor

```go
type Monitor interface {
	Observe(name string, value float64)
}
```

Monitor is used to monitor rothko's operation.

#### type Resources

```go
type Resources struct {
	Logger  Logger
	Monitor Monitor
}
```

Resources is a collection of all the external resources. It implements all of
the methods of the fields but in a nil-safe way.

```go
var Default Resources
```
Default is the default set of resources. Can be overridden by plugins.

#### func (Resources) Errorw

```go
func (r Resources) Errorw(msg string, keyvals ...interface{})
```
Errorw calls Logger.Errorw if Logger is not nil.

#### func (Resources) Infow

```go
func (r Resources) Infow(msg string, keyvals ...interface{})
```
Infow calls Logger.Infow if Logger is not nil.

#### func (Resources) Observe

```go
func (r Resources) Observe(name string, value float64)
```
Observe calls Monitor.Observe if Logger is not nil.
