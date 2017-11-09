// Copyright (C) 2017. See AUTHORS.

package draw

import (
	"testing"
)

func TestContext(t *testing.T) {
	cols := testMakeColumns(100, 30, 10, func(x, y int) float64 {
		return float64(x + y)
	})

	t.Run("Linear", func(t *testing.T) {
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

	b.Run("Linear", func(b *testing.B) {
		m := NewRGB(1000, 300)
		c := Context{
			Colors: grayscale,
			Canvas: m,
			Min:    30,
			Max:    100,
		}

		b.SetBytes(int64(len(m.Pix)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			c.Draw(cols)
		}
	})

	b.Run("Logarithm", func(b *testing.B) {
		m := NewRGB(1000, 300)
		c := Context{
			Colors:     grayscale,
			Canvas:     m,
			Min:        30,
			Max:        100,
			Logrithmic: true,
		}

		b.SetBytes(int64(len(m.Pix)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			c.Draw(cols)
		}
	})
}
