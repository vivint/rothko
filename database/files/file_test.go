// Copyright (C) 2018. See AUTHORS.

package files

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/spacemonkeygo/rothko/database/files/internal/meta"
	"github.com/spacemonkeygo/rothko/internal/assert"
)

// newTestFile constructs a temporary file backed by disk.
func newTestFile(t testing.TB) (f file, cleanup func()) {
	t.Helper()

	fh, err := ioutil.TempFile("", "file-")
	assert.NoError(t, err)
	assert.NoError(t, fh.Close())

	name := fh.Name()

	f, err = createFile(ctx, name, 512, 10)
	assert.NoError(t, err)

	return f, func() {
		f.Close()
		os.Remove(name)
	}
}

func TestFile(t *testing.T) {
	data := make([]byte, 100)
	for i := range data {
		data[i] = byte(i)
	}

	rec := record{
		version: recordVersion,
		kind:    recordKind_complete,
		start:   1234,
		end:     4567,
		size:    100,
		data:    data,
	}

	m := meta.Metadata{
		Size_: 512,
	}

	t.Run("Metadata", func(t *testing.T) {
		f, cleanup := newTestFile(t)
		defer cleanup()

		assert.NoError(t, f.SetMetadata(ctx, m))

		got, err := f.Metadata(ctx)
		assert.NoError(t, err)
		assert.DeepEqual(t, m, got)
	})

	t.Run("Basic", func(t *testing.T) {
		f, cleanup := newTestFile(t)
		defer cleanup()

		assert.NoError(t, f.SetRecord(ctx, 3, rec))

		got, err := f.Record(ctx, 3)
		assert.NoError(t, err)
		assert.DeepEqual(t, rec, got)
	})

	t.Run("HasRecord", func(t *testing.T) {
		f, cleanup := newTestFile(t)
		defer cleanup()

		assert.NoError(t, f.SetRecord(ctx, 3, rec))

		ok := f.HasRecord(ctx, 0)
		assert.That(t, !ok)

		ok = f.HasRecord(ctx, 3)
		assert.That(t, ok)
	})

	t.Run("OpenFails", func(t *testing.T) {
		fh, err := ioutil.TempFile("", "file-")
		assert.NoError(t, err)
		defer os.Remove(fh.Name())
		defer fh.Close()

		// no metadata
		_, err = openFile(ctx, fh.Name())
		assert.Error(t, err)

		assert.NoError(t, fh.Truncate(recordHeaderSize+100))

		// invalid metadata record
		_, err = openFile(ctx, fh.Name())
		assert.Error(t, err)
	})
}

func BenchmarkFile(b *testing.B) {
	data := make([]byte, 100)
	for i := range data {
		data[i] = byte(i)
	}

	rec := record{
		version: recordVersion,
		kind:    recordKind_complete,
		start:   1234,
		end:     4567,
		size:    100,
		data:    data,
	}

	m := meta.Metadata{
		Size_: 512,
	}

	b.Run("Metadata", func(b *testing.B) {
		b.Run("Write", func(b *testing.B) {
			f, cleanup := newTestFile(b)
			defer cleanup()

			b.ReportAllocs()
			defer b.StopTimer()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				f.SetMetadata(ctx, m)
			}
		})

		b.Run("Read", func(b *testing.B) {
			f, cleanup := newTestFile(b)
			defer cleanup()

			assert.NoError(b, f.SetMetadata(ctx, m))

			b.ReportAllocs()
			defer b.StopTimer()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				f.Metadata(ctx)
			}
		})
	})

	b.Run("Record", func(b *testing.B) {
		b.Run("Write", func(b *testing.B) {
			f, cleanup := newTestFile(b)
			defer cleanup()

			b.ReportAllocs()
			defer b.StopTimer()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				f.SetRecord(ctx, 3, rec)
			}
		})

		b.Run("Read", func(b *testing.B) {
			f, cleanup := newTestFile(b)
			defer cleanup()

			assert.NoError(b, f.SetRecord(ctx, 3, rec))

			b.ReportAllocs()
			defer b.StopTimer()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				f.Record(ctx, 3)
			}
		})

		b.Run("Has", func(b *testing.B) {
			f, cleanup := newTestFile(b)
			defer cleanup()

			assert.NoError(b, f.SetRecord(ctx, 3, rec))

			b.ReportAllocs()
			defer b.StopTimer()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				f.HasRecord(ctx, 3)
			}
		})
	})
}
