# package typeassert

`import "github.com/spacemonkeygo/rothko/internal/typeassert"`

package typeassert provides helper functions type asserting structures.

## Usage

#### type Asserter

```go
type Asserter struct {
}
```

Asserter helps type assertions with an "all-or-nothing" style API.

#### func  A

```go
func A(x interface{}) *Asserter
```
A wraps the value in an Asserter.

#### func (*Asserter) Bool

```go
func (a *Asserter) Bool() bool
```
Bool asserts the value as a bool.

#### func (*Asserter) Err

```go
func (a *Asserter) Err() error
```
Err returns an error if any of the assertions failed. If the error is not nil,
none of the assertions are valid.

#### func (*Asserter) Float64

```go
func (a *Asserter) Float64() float64
```
Float64 asserts the value as a float64.

#### func (*Asserter) I

```go
func (a *Asserter) I(index string) *Asserter
```
I indexes into a map[string]interface{}.

#### func (*Asserter) Int

```go
func (a *Asserter) Int() int
```
Int asserts the value as an int.

#### func (*Asserter) N

```go
func (a *Asserter) N(index int) *Asserter
```
N indexes into a []interface{}.

#### func (*Asserter) String

```go
func (a *Asserter) String() string
```
String asserts the value as a string.

#### func (*Asserter) V

```go
func (a *Asserter) V() interface{}
```
V returns the current value pointed at by the Asserter.
