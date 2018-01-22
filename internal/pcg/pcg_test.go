// Copyright (C) 2018. See AUTHORS.

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

func TestPCGFixed(t *testing.T) {
	pi := New(2345, 2378)

	out := make([]uint32, 10)
	for i := range out {
		out[i] = pi.Uint32()
	}

	assert.DeepEqual(t, out, []uint32{
		0xccca066b,
		0x40cee775,
		0x0df46902,
		0x981fbe29,
		0xfc8bfb85,
		0xcfd9eef2,
		0xa046c325,
		0x31abe14c,
		0xe29defb4,
		0x160568cc,
	})
}

// holds on to results in a way that the compiler won't remove
var blackHole PCG

func BenchmarkNewPCG(b *testing.B) {
	b.ReportAllocs()
	defer b.StopTimer()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		blackHole = New(uint64(i), 398)
	}
}

func BenchmarkPCG(b *testing.B) {
	p := New(42, 54)

	b.SetBytes(4)
	b.ReportAllocs()
	defer b.StopTimer()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p.Uint32()
	}
}
