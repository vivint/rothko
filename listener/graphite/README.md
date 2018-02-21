# package graphite

`import "github.com/vivint/rothko/listener/graphite"`

package graphite provides a listener for the graphite wire protocol.

## Usage

#### type Listener

```go
type Listener struct {
}
```

Listener implements the listener.Listener for the graphite wire protocol.

#### func  New

```go
func New(address string) *Listener
```
New returns a Listener that when Run will listen on the provided address.

#### func (*Listener) Run

```go
func (l *Listener) Run(ctx context.Context, w *data.Writer) (err error)
```
Run listens on the address and writes all of the metrics to the writer.
