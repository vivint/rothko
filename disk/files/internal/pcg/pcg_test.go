// Copyright (C) 2017. See AUTHORS.

package pcg

import (
	"testing"

	"github.com/spacemonkeygo/rothko/internal/assert"
)

func TestPCG(t *testing.T) {
	pi := New(0, 0)
	pz := PCG{}

	for i := 0; i < 10; i++ {
		assert.Equal(t, pi.Uint32(), pz.Uint32())
	}
}

func BenchmarkPCG(b *testing.B) {
	p := New(42, 54)

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		p.Uint32()
	}
}
