# package system

`import "github.com/spacemonkeygo/rothko/database/files/internal/system"`

package system provides optimized and dangerous functions for system calls.

## Usage

```go
const (
	PROT_READ  = syscall.PROT_READ
	PROT_WRITE = syscall.PROT_WRITE
	MAP_SHARED = syscall.MAP_SHARED
	MS_SYNC    = syscall.MS_SYNC
	MS_ASYNC   = syscall.MS_ASYNC
)
```

```go
var Error = errs.Class("system")
```

#### func  Allocate

```go
func Allocate(fd int, length int64) (err error)
```

#### func  Close

```go
func Close(fd uintptr) (err error)
```

#### func  Mmap

```go
func Mmap(fd int, length int, prot int, flags int) (data uintptr, err error)
```

#### func  Msync

```go
func Msync(data uintptr, length int, flags int) (err error)
```

#### func  Munmap

```go
func Munmap(data uintptr, length int) (err error)
```

#### func  NextDirent

```go
func NextDirent(buf []byte) (out_buf []byte, name []byte, ok bool)
```

#### func  Open

```go
func Open(path []byte) (fd uintptr, err error)
```
