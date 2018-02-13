# load
--
    import "github.com/spacemonkeygo/rothko/data/load"

package load provides a function to load a dist.Dist from a data.Record.

## Usage

#### func  Load

```go
func Load(ctx context.Context, rec data.Record) (dist.Dist, error)
```
Load returns the dist.Dist for the data.Record.
