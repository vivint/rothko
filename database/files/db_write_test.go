// Copyright (C) 2018. See AUTHORS.

package files

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/spacemonkeygo/rothko/internal/assert"
)

func TestDBWrite(t *testing.T) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	db, cleanup := newTestDB(t, Options{
		Size:  1024,
		Cap:   10,
		Files: 10,
	})
	defer cleanup()
	go db.Run(ctx)

	type res struct {
		b bool
		e error
	}

	ch := make(chan res)
	sendErr := func(ok bool, err error) { ch <- res{ok, err} }

	for i := 0; i < 200; i++ {
		db.Queue(ctx, "test.bar.baz", int64(i), int64(i+1),
			make([]byte, 700), sendErr)

		r := <-ch
		assert.That(t, r.b)
		assert.NoError(t, r.e)
	}
}

func BenchmarkDBWrite(b *testing.B) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// we have a really small cap here because HFS+ doesn't do sparse files.
	// gonna have to try it on linux!
	db, cleanup := newTestDB(b, Options{
		Size:  1024,   // 1K/record
		Cap:   102400, // 100MB/file
		Files: 10,     // 1GB/metric
	})
	defer cleanup()
	go db.Run(ctx)

	var mu sync.Mutex
	var ctr int

	b.ReportAllocs()
	defer b.StopTimer()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		mu.Lock()
		metric := fmt.Sprintf("test.bar.bench.%d", ctr)
		ctr++
		mu.Unlock()

		i := int64(0)
		for pb.Next() {
			db.Queue(ctx, metric, i, i+1, make([]byte, 300), nil)
			i++
		}
	})
}
