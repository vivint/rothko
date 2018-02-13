// Copyright (C) 2018. See AUTHORS.

package files

import (
	"context"
	"testing"

	"github.com/spacemonkeygo/rothko/internal/assert"
)

func TestDBRead(t *testing.T) {
	t.Run("Metrics", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		db, cleanup := newTestDB(t, Options{
			Size:  1024,
			Cap:   10,
			Files: 10,
		})
		defer cleanup()
		go db.Run(ctx)

		expected := testPopulateDB(t, db, 100)

		names := make(map[string]struct{})
		err := db.Metrics(ctx, func(name string) (ok bool, err error) {
			names[name] = struct{}{}
			return true, nil
		})
		assert.NoError(t, err)
		assert.DeepEqual(t, names, expected)
	})
}

func BenchmarkDBRead(b *testing.B) {
	b.Run("Metrics", func(b *testing.B) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		db, cleanup := newTestDB(b, Options{
			Size:  1024,
			Cap:   10,
			Files: 10,
		})
		defer cleanup()
		go db.Run(ctx)

		testPopulateDB(b, db, 100)

		// populate the metrics explicitly to avoid any background adding
		assert.NoError(b, db.PopulateMetrics(ctx))

		b.ReportAllocs()
		defer b.StopTimer()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			db.Metrics(ctx, func(name string) (ok bool, err error) {
				return true, nil
			})
		}
	})

	b.Run("MetricsContended", func(b *testing.B) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		db, cleanup := newTestDB(b, Options{
			Size:  1024,
			Cap:   10,
			Files: 10,
		})
		defer cleanup()
		go db.Run(ctx)

		testPopulateDB(b, db, 100)

		// populate the metrics explicitly to avoid any background adding
		assert.NoError(b, db.PopulateMetrics(ctx))

		b.ReportAllocs()
		defer b.StopTimer()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				db.Metrics(ctx, func(name string) (ok bool, err error) {
					return true, nil
				})
			}
		})
	})

	b.Run("PopulateMetrics", func(b *testing.B) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		db, cleanup := newTestDB(b, Options{
			Size:  1024,
			Cap:   10,
			Files: 10,
		})
		defer cleanup()
		go db.Run(ctx)

		testPopulateDB(b, db, 1000)

		b.ReportAllocs()
		defer b.StopTimer()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			assert.NoError(b, db.PopulateMetrics(ctx))
		}
	})
}
