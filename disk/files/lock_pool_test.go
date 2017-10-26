// Copyright (C) 2017. See AUTHORS.

package files

import (
	"testing"
)

func BenchmarkLockPool(b *testing.B) {
	l := newLockPool()
	key := randomMetric()

	b.ReportAllocs()
	defer b.StopTimer()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Lock(key)
		l.Unlock(key)
	}
}
