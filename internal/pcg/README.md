# pcg
--
    import "github.com/spacemonkeygo/rothko/internal/pcg"

package pcg provides the pcg random number generator

## Usage

#### type PCG

```go
type PCG struct {
}
```


#### func  New

```go
func New(state, inc uint64) PCG
```

#### func (*PCG) Uint32

```go
func (p *PCG) Uint32() uint32
```
