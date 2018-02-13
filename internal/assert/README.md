# package assert

`import "github.com/spacemonkeygo/rothko/internal/assert"`

package assert provides helper functions for tests.

## Usage

#### func  DeepEqual

```go
func DeepEqual(t testing.TB, a, b interface{})
```

#### func  Equal

```go
func Equal(t testing.TB, a, b interface{})
```

#### func  Error

```go
func Error(t testing.TB, err error)
```

#### func  Nil

```go
func Nil(t testing.TB, a interface{})
```

#### func  NoError

```go
func NoError(t testing.TB, err error)
```

#### func  NotNil

```go
func NotNil(t testing.TB, a interface{})
```

#### func  That

```go
func That(t testing.TB, v bool)
```
