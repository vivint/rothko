// Copyright (C) 2017. See AUTHORS.

package files

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"os"
	"testing"

	"github.com/spacemonkeygo/rothko/internal/assert"
)

// newTestMetric constructs a temporary metric.
func newTestMetric(t testing.TB) (m *metric, cleanup func()) {
	t.Helper()

	dir, err := ioutil.TempDir("", "metric-")
	assert.NoError(t, err)

	t.Log("temp dir:", dir)

	opts := metricOptions{
		fch: newFileCache(fileCacheOptions{
			Handles: 10,
			Size:    1024,
			Cap:     10,
		}),
		dir:  dir,
		name: "test.metric",
		max:  10,
	}

	m, err = newMetric(ctx, opts)
	assert.NoError(t, err)

	return m, func() {
		os.RemoveAll(dir)
	}
}

func TestMetric(t *testing.T) {
	t.Run("Write", func(t *testing.T) {
		m, cleanup := newTestMetric(t)
		defer cleanup()

		// test that a write that is too large cannot pass as the first write
		written, err := m.Write(ctx, 100, 200, make([]byte, 1024*1024))
		assert.Error(t, err)
		assert.That(t, !written)

		// test that a normal write works
		written, err = m.Write(ctx, 10, 20, make([]byte, 10))
		assert.NoError(t, err)
		assert.That(t, written)

		// test that a chronologically previous write does not work
		written, err = m.Write(ctx, 0, 10, make([]byte, 10))
		assert.NoError(t, err)
		assert.That(t, !written)

		// test that a write that is too large cannot pass after a valid write
		written, err = m.Write(ctx, 100, 200, make([]byte, 1024*1024))
		assert.Error(t, err)
		assert.That(t, !written)
	})

	t.Run("TimeRange", func(t *testing.T) {
		m, cleanup := newTestMetric(t)
		defer cleanup()

		for i := int64(0); i < 1000; i++ {
			written, err := m.Write(ctx, i, i+1, make([]byte, 10))
			assert.NoError(t, err)
			assert.That(t, written)

			first, last, err := m.TimeRange(ctx)
			assert.NoError(t, err)
			assert.Equal(t, last, i+1)
			// the timestamp of the earliest record will be 10 (the cap) times
			// the file index minus 1 (since files are 1 based)
			assert.Equal(t, first, int64(m.first-1)*10)
		}
	})

	t.Run("Search", func(t *testing.T) {
		m, cleanup := newTestMetric(t)
		defer cleanup()

		for i := int64(0); i < 1000; i++ {
			written, err := m.Write(ctx, 50*i, 50*i+20, make([]byte, 10))
			assert.NoError(t, err)
			assert.That(t, written)
		}

		// 890 because we can keep up to 110 records as there are 10 per file
		// and 10 files, and we have 1 file of staging data. everything before
		// the earliest record should be empty.
		for i := int64(-100); i < 890; i++ {
			num, head, err := m.Search(ctx, 50*i)
			assert.NoError(t, err)
			assert.Equal(t, num, 0)
			assert.Equal(t, head, 0)
		}

		// check right on the boundary and somewhere between records.
		for _, offset := range []int64{0, 10} {
			for i := int64(890); i < 1000; i++ {
				start := 50*i + offset
				num, head, err := m.Search(ctx, start)
				assert.NoError(t, err)

				rec, err := m.readRecord(ctx, num, head)
				assert.NoError(t, err)
				assert.That(t, rec.start <= start)

				// get the next record and make sure it's bigger
				if head+1 < 10 {
					rec, err = m.readRecord(ctx, num, head+1)
				} else if num < m.last {
					rec, err = m.readRecord(ctx, num+1, 0)
				} else {
					continue
				}
				assert.NoError(t, err)
				assert.That(t, rec.start > start)
			}
		}

		// everything after the last record should be the last record
		for i := int64(1000); i < 1100; i++ {
			num, head, err := m.Search(ctx, 50*i)
			assert.NoError(t, err)
			assert.Equal(t, num, 100)
			assert.Equal(t, head, 9)
		}
	})

	t.Run("Read", func(t *testing.T) {
		test := func(t *testing.T, buf_size int) {
			m, cleanup := newTestMetric(t)
			defer cleanup()

			for i := int64(0); i < 1000; i++ {
				buf := make([]byte, buf_size)
				binary.BigEndian.PutUint64(buf, uint64(i))

				written, err := m.Write(ctx, i, i+1, buf)
				assert.NoError(t, err)
				assert.That(t, written)
			}

			m.Read(ctx, 0, 10000, nil,
				func(start, end int64, data []byte) error {
					buf := make([]byte, buf_size)
					binary.BigEndian.PutUint64(buf, uint64(start))
					assert.That(t, bytes.Equal(data, buf))
					return nil
				})
		}

		t.Run("Small", func(t *testing.T) { test(t, 8) })
		t.Run("Large", func(t *testing.T) { test(t, 8+4096) })
	})

}

func BenchmarkMetric(b *testing.B) {
	b.Run("Write", func(b *testing.B) {
		m, cleanup := newTestMetric(b)
		defer cleanup()

		data := make([]byte, 100)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			m.Write(ctx, int64(i), int64(i+1), data)
		}
	})

	b.Run("TimeRange", func(b *testing.B) {
		m, cleanup := newTestMetric(b)
		defer cleanup()

		for i := int64(0); i < 1000; i++ {
			written, err := m.Write(ctx, i, i+1, make([]byte, 10))
			assert.NoError(b, err)
			assert.That(b, written)
		}

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			m.TimeRange(ctx)
		}
	})

	b.Run("Search", func(b *testing.B) {
		m, cleanup := newTestMetric(b)
		defer cleanup()

		for i := int64(0); i < 1000; i++ {
			written, err := m.Write(ctx, i, i+1, make([]byte, 10))
			assert.NoError(b, err)
			assert.That(b, written)
		}

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			m.Search(ctx, 890+int64(i%110))
		}
	})

	b.Run("Read", func(b *testing.B) {
		test := func(b *testing.B, buf_size int) {
			m, cleanup := newTestMetric(b)
			defer cleanup()

			for i := int64(0); i < 1000; i++ {
				buf := make([]byte, buf_size)
				binary.BigEndian.PutUint64(buf, uint64(i))

				written, err := m.Write(ctx, i, i+1, buf)
				assert.NoError(b, err)
				assert.That(b, written)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				m.Read(ctx, 0, 10000, nil,
					func(start, end int64, data []byte) error {
						return nil
					})
			}
		}

		b.Run("Small", func(b *testing.B) { test(b, 8) })
		b.Run("Large", func(b *testing.B) { test(b, 8+4096) })
	})
}
