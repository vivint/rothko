# query
--
    import "github.com/spacemonkeygo/rothko/api/query"

package query provides routines to query out of a set of metrics.

## Usage

#### type Search

```go
type Search struct {
}
```

Search represents a metric search.

#### func  New

```go
func New(query string, capacity int) *Search
```
New constructs a metric searcher from the query string.

#### func (*Search) Add

```go
func (s *Search) Add(name string) (bool, error)
```
Add is meant to be passed to a disk.Metrics call.

#### func (*Search) Match

```go
func (s *Search) Match(metric string) bool
```
Match checks if the Search matches the metric.

#### func (*Search) Matched

```go
func (s *Search) Matched() []string
```
Matched returns the matched metrics.
