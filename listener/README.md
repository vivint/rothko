# package listener

`import "github.com/spacemonkeygo/rothko/listener"`

package listener provides types for adding data to rothko.

## Usage

#### type Listener

```go
type Listener interface {
	// Run should Add values into the Writer until the context is canceled.
	Run(ctx context.Context, w *data.Writer) (err error)
}
```

Listener is a type that writes from some data source to the privided Writer.
