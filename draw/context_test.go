// Copyright (C) 2017. See AUTHORS.

package draw

import (
	"image/png"
	"os"
	"testing"

	"github.com/spacemonkeygo/rothko/internal/assert"
)

func TestContext(t *testing.T) {
	fh, err := os.OpenFile("test.png", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	assert.NoError(t, err)
	defer fh.Close()

	var cols []Column
	for i := 0; i < 1000; i++ {
		var data []float64
		for j := 0; j < 300; j++ {
			data = append(data, float64(i+j))
		}
		cols = append(cols, Column{
			X:    i,
			W:    1,
			Data: data,
		})
	}

	c := Context{
		Colors: dumb,
		Height: 300,
	}

	m := c.Draw(ctx, cols)
	assert.NoError(t, png.Encode(fh, m))
}

func BenchmarkContext(b *testing.B) {
	var cols []Column
	for i := 0; i < 1000; i++ {
		var data []float64
		for j := 0; j < 300; j++ {
			data = append(data, float64(i+j))
		}
		cols = append(cols, Column{
			X:    i,
			W:    1,
			Data: data,
		})
	}
	c := Context{
		Colors: dumb,
		Height: 300,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.Draw(ctx, cols)
	}
}
