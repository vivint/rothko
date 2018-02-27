// Copyright (C) 2018. See AUTHORS.

package graphite

import (
	"bytes"
	"context"
	"net"
	"strings"
	"testing"

	"github.com/vivint/rothko/data"
	"github.com/vivint/rothko/dist"
	"github.com/vivint/rothko/internal/assert"
)

func TestListener(t *testing.T) {
	ctx := context.Background()
	w := data.NewWriter(fakeParams{})

	lines := []byte(strings.Join([]string{
		"test.foo.bar 123 0",
		"test.foo.baz 123 0",
		"test.foo.bif 123 0",
		"test.foo.zoo 123 0",
	}, "\n"))

	assert.NoError(t, handleConn(ctx, w, newFakeConn(lines)))

	names := make(map[string]bool)
	w.Capture(ctx,
		func(ctx context.Context, name string, rec data.Record) bool {
			names[name] = true
			return true
		})

	assert.DeepEqual(t, names, map[string]bool{
		"test.foo.bar": true,
		"test.foo.baz": true,
		"test.foo.bif": true,
		"test.foo.zoo": true,
	})
}

//
// fakes. only required functions stubbed out. sorry if you break this
// accidentally!
//

type fakeConn struct {
	*bytes.Reader
	net.Conn
}

func newFakeConn(data []byte) *fakeConn {
	return &fakeConn{Reader: bytes.NewReader(data)}
}

func (fakeConn) Close() error                 { return nil }
func (fakeConn) RemoteAddr() net.Addr         { return &net.TCPAddr{} }
func (f fakeConn) Read(b []byte) (int, error) { return f.Reader.Read(b) }

type fakeParams struct{ dist.Params }

func (fakeParams) Kind() string            { return "fake" }
func (fakeParams) New() (dist.Dist, error) { return fakeDist{}, nil }

type fakeDist struct{ dist.Dist }

func (fakeDist) Kind() string            { return "fake" }
func (fakeDist) Observe(val float64)     {}
func (fakeDist) Marshal(x []byte) []byte { return x }
