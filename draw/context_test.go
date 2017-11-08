// Copyright (C) 2017. See AUTHORS.

package draw

import (
	"image/png"
	"os"
	"testing"

	"github.com/spacemonkeygo/rothko/internal/assert"
)

func debugWriteImage(t *testing.T, m *RGB) {
	fh, err := os.OpenFile("test.png", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	assert.NoError(t, err)
	defer fh.Close()
	assert.NoError(t, png.Encode(fh, m.AsImage()))
}

func TestContext(t *testing.T) {
	cols := testMakeColumns(100, 30, 10, func(x, y int) float64 {
		return float64(x + y)
	})

	t.Run("Value", func(t *testing.T) {
		m := NewRGB(1000, 300)
		c := Context{
			Colors: grayscale,
			Canvas: m,
			Min:    30,
			Max:    100,
		}
		c.Draw(cols)
	})

	t.Run("Logarithm", func(t *testing.T) {
		m := NewRGB(1000, 300)
		c := Context{
			Colors:     grayscale,
			Canvas:     m,
			Min:        30,
			Max:        100,
			Logrithmic: true,
		}
		c.Draw(cols)
	})
}

func BenchmarkContext(b *testing.B) {
	cols := testMakeColumns(100, 30, 10, func(x, y int) float64 {
		return float64(x + y)
	})
	m := NewRGB(1000, 300)
	c := Context{
		Colors: grayscale,
		Canvas: m,
		Min:    30,
		Max:    100,
	}

	b.SetBytes(int64(4 * m.Width * m.Height))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.Draw(cols)
	}
}
