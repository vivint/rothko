// Copyright (C) 2018. See AUTHORS.

package data

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/spacemonkeygo/rothko/internal/assert"
	"github.com/spacemonkeygo/rothko/internal/pcg"
)

func TestAgg(t *testing.T) {
	var params fakeParams
	a := newAgg(params, time.Now())

	for _, i := range rand.Perm(10) {
		a.Observe(float64(i), []byte(fmt.Sprint(i)))
	}

	_, rec := a.Finish(nil, time.Now())

	assert.Equal(t, rec.Min, float64(0))
	assert.Equal(t, string(rec.MinId), "0")
	assert.Equal(t, rec.Max, float64(9))
	assert.Equal(t, string(rec.MaxId), "9")
	assert.That(t, rec.StartTime < rec.EndTime)
	assert.Equal(t, rec.Kind, params.Kind())
	assert.That(t, len(rec.Distribution) > 0)
}

func BenchmarkAgg(b *testing.B) {
	a := newAgg(fakeParams{}, time.Now())

	b.ReportAllocs()
	defer b.StopTimer()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var rng pcg.PCG
		for pb.Next() {
			a.Observe(float64(rng.Uint32()), []byte("some id"))
		}
	})
}
