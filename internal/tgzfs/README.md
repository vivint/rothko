# package tgzfs

`import "github.com/spacemonkeygo/rothko/internal/tgzfs"`

package tgzfs provides an http.FileSystem based on a tgz.

## Usage

#### type FS

```go
type FS struct {
}
```

FS is an http.FileServer for a tarball.

#### func  New

```go
func New(data []byte) (*FS, error)
```
New constructs a FS from a gzip encoded tar ball in the data.

#### func (*FS) Open

```go
func (fs *FS) Open(name string) (http.File, error)
```
Open returns an http.File for the given path.
