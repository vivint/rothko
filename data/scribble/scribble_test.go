// Copyright (C) 2017. See AUTHORS.

package scribble

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/spacemonkeygo/rothko/data"
	"github.com/spacemonkeygo/rothko/data/dists/tdigest"
	"github.com/spacemonkeygo/rothko/internal/assert"
)

func TestScribble(t *testing.T) {
	s := NewScribbler(tdigest.Params{
		Compression: 10,
	})

	s.Scribble(ctx, "1", 1, nil)
	s.Scribble(ctx, "2", 2, nil)
	s.Scribble(ctx, "3", 3, nil)
	s.Scribble(ctx, "4", 4, nil)

	got := make(map[string]bool)
	s.Capture(ctx, func(metric string, rec data.Record) bool {
		got[metric] = true
		return true
	})

	assert.That(t, got["1"])
	assert.That(t, got["2"])
	assert.That(t, got["3"])
	assert.That(t, got["4"])

	got = make(map[string]bool)
	s.Capture(ctx, func(metric string, rec data.Record) bool {
		got[metric] = true
		return true
	})

	assert.That(t, len(got) == 0)
}

func BenchmarkScribble(b *testing.B) {
	s := NewScribbler(tdigest.Params{
		Compression: 10,
	})

	metrics := make([]string, 16)
	for i := range metrics {
		metrics[i] = fmt.Sprintf("metric%d", i)
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		skip := rand.Intn(16)
		for i := 0; pb.Next(); i += skip {
			s.Scribble(ctx, metrics[i&15], 1.65, []byte("some id"))
		}
	})
}
