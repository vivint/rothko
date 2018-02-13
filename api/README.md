# package api

`import "github.com/spacemonkeygo/rothko/api"`

package api provides apis for interacting with a rothko server

## Usage

#### type Server

```go
type Server struct {
}
```

Server is an http.Handler that can serve responses for a frontend.

#### func  New

```go
func New(db database.DB) *Server
```
New returns a new Server.

#### func (*Server) ServeHTTP

```go
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request)
```
ServeHTTP implements the http.Handler interface for the server. It just looks at
the method and last path component to route.
