# package api

`import "github.com/vivint/rothko/api"`

package api provides apis for interacting with a rothko server

## Usage

#### type Options

```go
type Options struct {
	// Origin is sent back in Access-Control-Allow-Origin. If not set, sends
	// back '*'.
	Origin string

	// Username and Password control basic auth to the server. If unset, no
	// basic auth will be required.
	Username string
	Password string
}
```

Options for the server.

#### type Server

```go
type Server struct {
}
```

Server is an http.Handler that can serve responses for a frontend.

#### func  New

```go
func New(db database.DB, static http.Handler, opts Options) *Server
```
New returns a new Server.

#### func (*Server) ServeHTTP

```go
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request)
```
ServeHTTP implements the http.Handler interface for the server. It just looks at
the method and last path component to route.
