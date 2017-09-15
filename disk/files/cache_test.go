// Copyright (C) 2017. See AUTHORS.

package files

import (
	"testing"

	"github.com/spacemonkeygo/rothko/internal/assert"
)

// newTestCacheFile creates a file for the cache tests and will cause all sorts
// of trouble if used, but they should be safe to close. they also have the
// property that they can be compared for equality.
func newTestCacheFile(n int) file {
	return file{len: n}
}

func TestCache(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		ch := newCache(10)

		tok := ch.Put(newTestCacheFile(1))
		got, ok := ch.Take(tok)
		assert.That(t, ok)
		assert.Equal(t, got, newTestCacheFile(1))

		got, ok = ch.Take(tok)
		assert.That(t, !ok)
	})

	t.Run("Evict", func(t *testing.T) {
		ch := newCache(2)

		tok1 := ch.Put(newTestCacheFile(1))
		tok2 := ch.Put(newTestCacheFile(2))
		tok3 := ch.Put(newTestCacheFile(3))

		got, ok := ch.Take(tok1)
		assert.That(t, ok)
		assert.Equal(t, got, newTestCacheFile(1))

		got, ok = ch.Take(tok2)
		assert.That(t, !ok)

		got, ok = ch.Take(tok3)
		assert.That(t, ok)
		assert.Equal(t, got, newTestCacheFile(3))
	})
}

func BenchmarkCache(b *testing.B) {
	b.Run("PutTake", func(b *testing.B) {
		ch := newCache(10)
		f := newTestCacheFile(1)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			ch.Take(ch.Put(f))
		}
	})
}
