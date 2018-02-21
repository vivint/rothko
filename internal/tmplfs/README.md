# package tmplfs

`import "github.com/vivint/rothko/internal/tmplfs"`

package tmplfs wraps an http.FileSystem to make html files html/templates.

## Usage

#### type FS

```go
type FS struct {
}
```

FS wraps a http.FileSystem to add templates to files that end in .html.

#### func  New

```go
func New(fs http.FileSystem) *FS
```
New constructs a FS around the http.FileSystem. It implements http.Handler.

#### func (*FS) ServeHTTP

```go
func (s *FS) ServeHTTP(w http.ResponseWriter, req *http.Request)
```
ServeHTTP conforms to the http.Handler interface.
