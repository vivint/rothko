// Copyright (C) 2018. See AUTHORS.

package data

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/spacemonkeygo/rothko/internal/assert"
)

func TestWriter(t *testing.T) {
	ctx := context.Background()

	w := NewWriter(fakeParams{})

	w.Add(ctx, "1", 1, nil)
	w.Add(ctx, "2", 2, nil)
	w.Add(ctx, "3", 3, nil)
	w.Add(ctx, "4", 4, nil)

	got := make(map[string]bool)
	w.Iterate(ctx, func(ctx context.Context, metric string, rec Record) bool {
		got[metric] = true
		return true
	})

	assert.That(t, got["1"])
	assert.That(t, got["2"])
	assert.That(t, got["3"])
	assert.That(t, got["4"])

	got = make(map[string]bool)
	w.Capture(ctx, func(ctx context.Context, metric string, rec Record) bool {
		got[metric] = true
		return true
	})

	assert.That(t, got["1"])
	assert.That(t, got["2"])
	assert.That(t, got["3"])
	assert.That(t, got["4"])

	got = make(map[string]bool)
	w.Capture(ctx, func(ctx context.Context, metric string, rec Record) bool {
		got[metric] = true
		return true
	})

	assert.That(t, len(got) == 0)
}

func BenchmarkWriter(b *testing.B) {
	ctx := context.Background()

	b.Run("Add", func(b *testing.B) {
		w := NewWriter(fakeParams{})

		metrics := make([]string, 16)
		for i := range metrics {
			metrics[i] = fmt.Sprintf("metric%d", i)
		}

		b.ReportAllocs()
		defer b.StopTimer()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			skip := rand.Intn(16)
			for i := 0; pb.Next(); i += skip {
				w.Add(ctx, metrics[i&15], 1.65, []byte("some id"))
			}
		})
	})

	b.Run("Iterate", func(b *testing.B) {
		w := NewWriter(fakeParams{})

		bytes := int64(0)
		iterate := func(ctx context.Context, metric string, rec Record) bool {
			bytes = int64(len(rec.Distribution))
			return true
		}

		for i := 0; i < 100; i++ {
			w.Add(ctx, "metric", float64(i), nil)
		}

		b.ReportAllocs()
		defer b.StopTimer()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			w.Iterate(ctx, iterate)
		}

		b.SetBytes(bytes)
	})
}
