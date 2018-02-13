# package junk

`import "github.com/spacemonkeygo/rothko/internal/junk"`

package junk is a drop spot for things with no home.

## Usage

#### type Launcher

```go
type Launcher struct {
}
```


#### func (*Launcher) Queue

```go
func (l *Launcher) Queue(fn func(ctx context.Context, errch chan error))
```

#### func (*Launcher) Run

```go
func (l *Launcher) Run(ctx context.Context) error
```

#### type Tabbed

```go
type Tabbed struct {
}
```


#### func  NewTabbed

```go
func NewTabbed(w io.Writer) *Tabbed
```

#### func (*Tabbed) Error

```go
func (t *Tabbed) Error() error
```

#### func (*Tabbed) Flush

```go
func (t *Tabbed) Flush()
```

#### func (*Tabbed) Write

```go
func (t *Tabbed) Write(values ...string)
```
