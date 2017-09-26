// Copyright (C) 2017. See AUTHORS.

package files

import (
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

	m, err = newMetric(opts)
	assert.NoError(t, err)

	return m, func() {
		os.RemoveAll(dir)
	}
}

func TestMetric(t *testing.T) {
	m, cleanup := newTestMetric(t)
	defer cleanup()

	// test that a write that is too large cannot pass as the first write
	written, err := m.write(ctx, 100, 200, make([]byte, 1024*1024))
	assert.Error(t, err)
	assert.That(t, !written)

	// test that a normal write works
	written, err = m.write(ctx, 10, 20, make([]byte, 10))
	assert.NoError(t, err)
	assert.That(t, written)

	// test that a chronologically previous write does not work
	written, err = m.write(ctx, 0, 10, make([]byte, 10))
	assert.NoError(t, err)
	assert.That(t, !written)

	// test that a write that is too large cannot pass after a valid write
	written, err = m.write(ctx, 100, 200, make([]byte, 1024*1024))
	assert.Error(t, err)
	assert.That(t, !written)
}
