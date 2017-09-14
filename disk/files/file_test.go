// Copyright (C) 2017. See AUTHORS.

package files

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/spacemonkeygo/rothko/internal/assert"
)

func TestFileBasic(t *testing.T) {
	fh, err := ioutil.TempFile("", "file-basic-")
	assert.NoError(t, err)
	fh.Close()
	defer os.Remove(fh.Name())

	f, err := open(fh.Name(), 512)
	assert.NoError(t, err)
	defer f.Close()

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

	assert.NoError(t, f.put(3, rec))

	got, err := f.get(3)
	assert.NoError(t, err)
	assert.DeepEqual(t, rec, got)
}
