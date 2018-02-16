# package static

`import "github.com/spacemonkeygo/rothko/api/static"`

package static provides helpers for the static file server.

## Usage

#### type Static

```go
type Static struct {
}
```

Static wraps a http.FileSystem to add templates to files that end in .html.

#### func  New

```go
func New(fs http.FileSystem) *Static
```
New constructs a Static around the http.FileSystem.

#### func (*Static) ServeHTTP

```go
func (s *Static) ServeHTTP(w http.ResponseWriter, req *http.Request)
```
ServeHTTP conforms to the http.Handler interface.
