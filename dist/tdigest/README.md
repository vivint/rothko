# package tdigest

`import "github.com/spacemonkeygo/rothko/dist/tdigest"`

package tdigest provides a wrapper around github.com/zeebo/tdigest.

## Usage

#### type Params

```go
type Params struct {
	Compression float64
}
```

Params implements dist.Params for a t-digest distribution.

#### func (Params) Kind

```go
func (p Params) Kind() string
```
Kind returns the TDigest distribution kind.

#### func (Params) New

```go
func (p Params) New() (dist.Dist, error)
```
New returns a new TDigest as a dist.Dist.

#### func (Params) Unmarshal

```go
func (p Params) Unmarshal(data []byte) (dist.Dist, error)
```
Unmarshal loads a dist.Dist out of some bytes.

#### type Wrapper

```go
type Wrapper struct {
}
```

Wrapper implements dist.Dist for a t-digest.

#### func  Wrap

```go
func Wrap(td *tdigest.TDigest) *Wrapper
```
Wrap wraps the given t-digest.

#### func (*Wrapper) CDF

```go
func (w *Wrapper) CDF(x float64) float64
```
CDF returns the estimate CDF at the value x.

#### func (Wrapper) Kind

```go
func (Wrapper) Kind() string
```
Kind returns the string "tdigest".

#### func (Wrapper) Len

```go
func (w Wrapper) Len() int64
```
Len returns how many items were added to the t-digest.

#### func (Wrapper) Marshal

```go
func (w Wrapper) Marshal(buf []byte) []byte
```
Marshal appends a byte form of the t-digest to the provided buffer.

#### func (Wrapper) Observe

```go
func (w Wrapper) Observe(val float64)
```
Observe adds the value to the t-digest.

#### func (Wrapper) Query

```go
func (w Wrapper) Query(x float64) float64
```
Query returns the approximate x'th quantile.

#### func (Wrapper) Underlying

```go
func (w Wrapper) Underlying() *tdigest.TDigest
```
Underlying returns the underlying t-digest.
