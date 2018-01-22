// Copyright (C) 2018. See AUTHORS.

package heatmap

import (
	"testing"

	"github.com/spacemonkeygo/rothko/draw"
)

func TestContext(t *testing.T) {
	cols, linear := testMakeColumns(100, 30, 10, func(x, y int) float64 {
		return float64(x + y)
	})

	t.Run("Draw", func(t *testing.T) {
		m := draw.NewRGB(1000, 300)
		d := New(Options{
			Colors: grayscale,
			Canvas: m,
			Map:    linear,
		})
		for _, col := range cols {
			d.Draw(col)
		}
	})
}

func BenchmarkContext(b *testing.B) {
	cols, linear := testMakeColumns(100, 30, 10, func(x, y int) float64 {
		return float64(x + y)
	})

	b.Run("Draw", func(b *testing.B) {
		m := draw.NewRGB(1000, 300)
		d := New(Options{
			Colors: grayscale,
			Canvas: m,
			Map:    linear,
		})

		b.SetBytes(int64(len(m.Pix)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			for _, col := range cols {
				d.Draw(col)
			}
		}
	})
}
