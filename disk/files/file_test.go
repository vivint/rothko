// Copyright (C) 2017. See AUTHORS.

package files

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/spacemonkeygo/rothko/disk/files/internal/meta"
	"github.com/spacemonkeygo/rothko/internal/assert"
)

func newTestFile(t *testing.T) (f file, cleanup func()) {
	t.Helper()

	fh, err := ioutil.TempFile("", "file-")
	assert.NoError(t, err)
	assert.NoError(t, fh.Close())

	name := fh.Name()

	f, err = create(name, 512, 0)
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
		Head:  5,
	}

	t.Run("Metadata", func(t *testing.T) {
		f, cleanup := newTestFile(t)
		defer cleanup()

		assert.NoError(t, f.SetMetadata(m))

		got, err := f.Metadata()
		assert.NoError(t, err)
		assert.DeepEqual(t, m, got)
	})

	t.Run("Basic", func(t *testing.T) {
		f, cleanup := newTestFile(t)
		defer cleanup()

		assert.NoError(t, f.SetRecord(3, rec))

		got, err := f.Record(3)
		assert.NoError(t, err)
		assert.DeepEqual(t, rec, got)
	})

	t.Run("OpenFails", func(t *testing.T) {
		fh, err := ioutil.TempFile("", "file-")
		assert.NoError(t, err)
		defer os.Remove(fh.Name())
		defer fh.Close()

		// no metadata
		_, err = open(fh.Name())
		assert.Error(t, err)

		assert.NoError(t, fh.Truncate(recordHeaderSize+100))

		// invalid metadata record
		_, err = open(fh.Name())
		assert.Error(t, err)
	})
}
