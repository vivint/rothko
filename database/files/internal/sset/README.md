# package sset

`import "github.com/spacemonkeygo/rothko/database/files/internal/sset"`

package sset implements a sorted set of strings.

## Usage

#### type Set

```go
type Set struct {
}
```

Set represents an ordered set of strings.

#### func  New

```go
func New(cap int) *Set
```
New constructs a Set with the given initial capacity.

#### func (*Set) Add

```go
func (s *Set) Add(x string)
```

#### func (*Set) Copy

```go
func (s *Set) Copy() *Set
```
Copy returns a copy of the set.

#### func (*Set) Has

```go
func (s *Set) Has(x string) bool
```
Has returns if the set has the key.

#### func (*Set) Iter

```go
func (s *Set) Iter(cb func(name string) bool)
```
Iter iterates over all of the keys in the set.

#### func (*Set) Len

```go
func (s *Set) Len() int
```
Len returns the amount of elements in the set.

#### func (*Set) Merge

```go
func (s *Set) Merge(o *Set)
```
Merge inserts all of the values in o into the set.
