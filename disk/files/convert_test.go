// Copyright (C) 2017. See AUTHORS.

package files

import (
	"testing"

	"github.com/spacemonkeygo/rothko/internal/assert"
)

func TestMetricToPath(t *testing.T) {
	assert.Equal(t, metricToPath(`foo.bar`), `foo/bar.data`)
	assert.Equal(t, metricToPath(`foo..bar`), `foo/.bar.data`)
	assert.Equal(t, metricToPath(`foo....bar`), `foo/...bar.data`)
	assert.Equal(t, metricToPath(`fo/o.bar`), `fo%2fo/bar.data`)
	assert.Equal(t, metricToPath(`fo%o.bar`), `fo%25o/bar.data`)
	assert.Equal(t, metricToPath(`foo.bar.baz`), `foo/bar/baz.data`)
	assert.Equal(t, metricToPath(`foo.bar.`), `foo/bar/.data`)
	assert.Equal(t, metricToPath(`foo.bar..`), `foo/bar/..data`)
}
