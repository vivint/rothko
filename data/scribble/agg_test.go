// Copyright (C) 2017. See AUTHORS.

package scribble

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/spacemonkeygo/rothko/data/dists/tdigest"
	"github.com/spacemonkeygo/rothko/internal/assert"
)

func TestAgg(t *testing.T) {
	params := tdigest.Params{Compression: 10}
	a := newAgg(params, time.Now())

	for _, i := range rand.Perm(10) {
		a.Observe(float64(i), []byte(fmt.Sprint(i)))
	}

	rec := a.Finish(time.Now())

	assert.Equal(t, rec.Min, float64(0))
	assert.Equal(t, string(rec.MinId), "0")
	assert.Equal(t, rec.Max, float64(9))
	assert.Equal(t, string(rec.MaxId), "9")
	assert.That(t, rec.StartTime < rec.EndTime)
	assert.Equal(t, rec.DistributionKind, params.Kind())
	assert.That(t, len(rec.Distribution) > 0)
}

func BenchmarkAgg(b *testing.B) {
	params := tdigest.Params{Compression: 10}
	a := newAgg(params, time.Now())

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			a.Observe(float64(rng.Int63()), []byte("some id"))
		}
	})
}
