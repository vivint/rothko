// Copyright (C) 2017. See AUTHORS.

package files

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"testing"

	"github.com/spacemonkeygo/rothko/internal/assert"
)

// newTestDB constructs a temporary db.
func newTestDB(t testing.TB, opts Options) (db *DB, cleanup func()) {
	t.Helper()

	dir, err := ioutil.TempDir("", "db-")
	assert.NoError(t, err)

	t.Log("temp dir:", dir)

	return New(dir, opts), func() {
		os.RemoveAll(dir)
	}
}

func TestDBWrite(t *testing.T) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	db, cleanup := newTestDB(t, Options{
		Size:  512,
		Cap:   10,
		Files: 10,
	})
	defer cleanup()
	go db.Run(ctx)

	ch := make(chan error)
	db.QueueCB(ctx, "test.bar.baz", 0, 1, make([]byte, 700), func(err error) {
		ch <- err
	})
	assert.NoError(t, <-ch)
}

func BenchmarkDBWrite(b *testing.B) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// we have a really small cap here because HFS+ doesn't do sparse files.
	// gonna have to try it on linux!
	db, cleanup := newTestDB(b, Options{
		Size:  512,
		Cap:   1024,
		Files: 10,
	})
	defer cleanup()
	go db.Run(ctx)

	b.ReportAllocs()
	b.ResetTimer()

	var mu sync.Mutex
	var ctr int

	b.RunParallel(func(pb *testing.PB) {
		mu.Lock()
		metric := fmt.Sprintf("test.bar.bench.%d", ctr)
		ctr++
		mu.Unlock()

		for pb.Next() {
			db.Queue(ctx, metric, 0, 1, make([]byte, 300))
		}
	})
}
