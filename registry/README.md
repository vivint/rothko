# package registry

`import "github.com/vivint/rothko/registry"`

package registry provides ways for plugins to add rothko entities.

## Usage

#### func  NewDatabase

```go
func NewDatabase(ctx context.Context, name string, config interface{}) (
	database.DB, error)
```
NewDatabase constructs a Database using the DatabaseMaker registered under the
name in the Default registry. It returns an error if there has been no such
registration.

#### func  NewDistribution

```go
func NewDistribution(ctx context.Context, name string, config interface{}) (
	dist.Params, error)
```
NewDistribution constructs a Distribution using the DistributionMaker registered
under the name in the Default registry. It returns an error if there has been no
such registration.

#### func  NewListener

```go
func NewListener(ctx context.Context, name string, config interface{}) (
	listener.Listener, error)
```
NewListener constructs a Listener using the ListenerMaker registered under the
name in the Default registry. It returns an error if there has been no such
registration.

#### func  RegisterDatabase

```go
func RegisterDatabase(name string, maker DatabaseMaker)
```
RegisterDatabase registers the DatabaseMaker as the provided name in the Default
registry. It overwrites any previous calls for the same name.

#### func  RegisterDistribution

```go
func RegisterDistribution(name string, maker DistributionMaker)
```
RegisterDistribution registers the DistributionMaker as the provided name in the
Default registry. It overwrites any previous calls for the same name.

#### func  RegisterListener

```go
func RegisterListener(name string, maker ListenerMaker)
```
RegisterListener registers the ListenerMaker as the provided name in the Default
registry. It overwrites any previous calls for the same name.

#### type DatabaseMaker

```go
type DatabaseMaker interface {
	New(ctx context.Context, config interface{}) (database.DB, error)
}
```

DatabaseMaker constructs a listener from the provided config.

#### type DatabaseMakerFunc

```go
type DatabaseMakerFunc func(context.Context, interface{}) (database.DB, error)
```

DatabaseMakerFunc is a function type that implements DatabaseMaker.

#### func (DatabaseMakerFunc) New

```go
func (fn DatabaseMakerFunc) New(ctx context.Context, config interface{}) (
	database.DB, error)
```
New calls the DatabaseMakerFunc.

#### type DistributionMaker

```go
type DistributionMaker interface {
	New(ctx context.Context, config interface{}) (dist.Params, error)
}
```

DistributionMaker constructs a listener from the provided config.

#### type DistributionMakerFunc

```go
type DistributionMakerFunc func(context.Context, interface{}) (
	dist.Params, error)
```

DistributionMakerFunc is a function type that implements DistributionMaker.

#### func (DistributionMakerFunc) New

```go
func (fn DistributionMakerFunc) New(ctx context.Context, config interface{}) (
	dist.Params, error)
```
New calls the DistributionMakerFunc.

#### type ListenerMaker

```go
type ListenerMaker interface {
	New(ctx context.Context, config interface{}) (listener.Listener, error)
}
```

ListenerMaker constructs a listener from the provided config.

#### type ListenerMakerFunc

```go
type ListenerMakerFunc func(context.Context, interface{}) (
	listener.Listener, error)
```

ListenerMakerFunc is a function type that implements ListenerMaker.

#### func (ListenerMakerFunc) New

```go
func (fn ListenerMakerFunc) New(ctx context.Context, config interface{}) (
	listener.Listener, error)
```
New calls the ListenerMakerFunc.

#### type Registry

```go
type Registry struct {
}
```

Registry keeps track of a set of Makers by name.

```go
var Default Registry
```
Default is the default registry that the Register calls insert into.

#### func (*Registry) NewDatabase

```go
func (r *Registry) NewDatabase(ctx context.Context, name string,
	config interface{}) (database.DB, error)
```
NewDatabase constructs a Database using the DatabaseMaker registered under the
name. It returns an error if there has been no such registration.

#### func (*Registry) NewDistribution

```go
func (r *Registry) NewDistribution(ctx context.Context, name string,
	config interface{}) (dist.Params, error)
```
NewDistribution constructs a Distribution using the DistributionMaker registered
under the name. It returns an error if there has been no such registration.

#### func (*Registry) NewListener

```go
func (r *Registry) NewListener(ctx context.Context, name string,
	config interface{}) (listener.Listener, error)
```
NewListener constructs a Listener using the ListenerMaker registered under the
name. It returns an error if there has been no such registration.

#### func (*Registry) RegisterDatabase

```go
func (r *Registry) RegisterDatabase(name string, maker DatabaseMaker)
```
RegisterDatabase registers the DatabaseMaker under the given name. It overwrites
any previous calls for the same name.

#### func (*Registry) RegisterDistribution

```go
func (r *Registry) RegisterDistribution(name string, maker DistributionMaker)
```
RegisterDistribution registers the DistributionMaker under the given name. It
overwrites any previous calls for the same name.

#### func (*Registry) RegisterListener

```go
func (r *Registry) RegisterListener(name string, maker ListenerMaker)
```
RegisterListener registers the ListenerMaker under the given name. It overwrites
any previous calls for the same name.
