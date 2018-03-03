# package ui

`import "github.com/vivint/rothko/ui"`

package ui provides a gzipped tar archive of the compiled ui.

## Usage

```go
var Tarball []byte
```
Tarball contains a gzipped tar archive to be served for the ui. If it is nil, no
ui is served. If you want it to not be nil, use `roth generate`.
