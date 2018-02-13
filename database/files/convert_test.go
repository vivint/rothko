// Copyright (C) 2018. See AUTHORS.

package files

import (
	"testing"

	"github.com/spacemonkeygo/rothko/internal/assert"
)

func TestConvert(t *testing.T) {
	t.Run("MetricToDir", func(t *testing.T) {
		f := func(metric string) string {
			return string(metricToDir(nil, metric))
		}

		assert.Equal(t, f(`foo.bar`), `foo/bar`)
		assert.Equal(t, f(`foo..bar`), `foo/%2ebar`)
		assert.Equal(t, f(`foo....bar`), `foo/%2e%2e%2ebar`)
		assert.Equal(t, f(`fo/o.bar`), `fo%2fo/bar`)
		assert.Equal(t, f(`fo%o.bar`), `fo%25o/bar`)
		assert.Equal(t, f(`foo.bar.baz`), `foo/bar/baz`)
		assert.Equal(t, f(`foo.bar.`), `foo/bar/%2e`)
		assert.Equal(t, f(`foo.bar..`), `foo/bar/%2e%2e`)
		assert.Equal(t, f(``), ``)
		assert.Equal(t, f(`.`), `%2e`)
		assert.Equal(t, f(`...`), `%2e%2e%2e`)
		assert.Equal(t, f(`.foo.bar`), `%2e/foo/bar`)
		assert.Equal(t, f(`...foo.bar`), `%2e%2e%2e/foo/bar`)
	})

	t.Run("DirToMetric", func(t *testing.T) {
		f := func(dir string) string {
			t.Helper()
			buf, err := dirToMetric(nil, []byte(dir))
			assert.NoError(t, err)
			return string(buf)
		}

		assert.Equal(t, f(`foo/bar`), `foo.bar`)
		assert.Equal(t, f(`foo/%2ebar`), `foo..bar`)
		assert.Equal(t, f(`foo/%2e%2e%2ebar`), `foo....bar`)
		assert.Equal(t, f(`fo%2fo/bar`), `fo/o.bar`)
		assert.Equal(t, f(`fo%25o/bar`), `fo%o.bar`)
		assert.Equal(t, f(`foo/bar/baz`), `foo.bar.baz`)
		assert.Equal(t, f(`foo/bar/%2e`), `foo.bar.`)
		assert.Equal(t, f(`foo/bar/%2e%2e`), `foo.bar..`)
		assert.Equal(t, f(``), ``)
		assert.Equal(t, f(`%2e`), `.`)
		assert.Equal(t, f(`%2e%2e%2e`), `...`)
		assert.Equal(t, f(`%2e/foo/bar`), `.foo.bar`)
		assert.Equal(t, f(`%2e%2e%2e/foo/bar`), `...foo.bar`)
	})

	t.Run("MetricToPath", func(t *testing.T) {
		f := func(metric string, num int) string {
			return string(metricToPath(nil, metric, num))
		}

		assert.Equal(t, f(`foo.bar`, 0), `foo/bar/0.data`)
		assert.Equal(t, f(`foo..bar`, 1), `foo/%2ebar/1.data`)
		assert.Equal(t, f(`foo....bar`, 2), `foo/%2e%2e%2ebar/2.data`)
		assert.Equal(t, f(`fo/o.bar`, 3), `fo%2fo/bar/3.data`)
		assert.Equal(t, f(`fo%o.bar`, 4), `fo%25o/bar/4.data`)
		assert.Equal(t, f(`foo.bar.baz`, 5), `foo/bar/baz/5.data`)
		assert.Equal(t, f(`foo.bar.`, 6), `foo/bar/%2e/6.data`)
		assert.Equal(t, f(`foo.bar..`, 7), `foo/bar/%2e%2e/7.data`)
		assert.Equal(t, f(``, 0), `0.data`)
		assert.Equal(t, f(`.`, 0), `%2e/0.data`)
		assert.Equal(t, f(`...`, 0), `%2e%2e%2e/0.data`)
		assert.Equal(t, f(`.foo.bar`, 0), `%2e/foo/bar/0.data`)
		assert.Equal(t, f(`...foo.bar`, 0), `%2e%2e%2e/foo/bar/0.data`)
	})
}

func BenchmarkConvert(b *testing.B) {
	b.Run("MetricToDir", func(b *testing.B) {
		var buf []byte

		b.ReportAllocs()
		defer b.StopTimer()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			buf = metricToDir(buf[:0], "some.stinking.metric")
		}

		b.SetBytes(int64(len(buf)))
	})

	b.Run("DirToMetric", func(b *testing.B) {
		var buf []byte
		dir := []byte("some/stinking/metric")

		b.ReportAllocs()
		defer b.StopTimer()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			buf, _ = dirToMetric(buf[:0], dir)
		}

		b.SetBytes(int64(len(buf)))
	})

	b.Run("MetricToPath", func(b *testing.B) {
		var buf []byte

		b.ReportAllocs()
		defer b.StopTimer()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			buf = metricToPath(buf[:0], "some.stinking.metric", 10)
		}

		b.SetBytes(int64(len(buf)))
	})
}
