// Copyright (C) 2017. See AUTHORS.

package draw

import (
	"image"
	"image/png"
	"os"
	"testing"

	"github.com/spacemonkeygo/rothko/internal/assert"
)

var grayscale []Color

func init() {
	for i := 0; i < 256; i++ {
		grayscale = append(grayscale, Color{
			R: uint8(i), G: uint8(i), B: uint8(i),
		})
	}
}

func (m *RGB) AsImage() *image.RGBA {
	return &image.RGBA{
		Pix:    m.Pix,
		Stride: m.Stride,
		Rect:   image.Rect(0, 0, m.Width, m.Height),
	}
}

func debugWriteImage(t *testing.T, m *RGB) {
	fh, err := os.OpenFile("test.png", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	assert.NoError(t, err)
	defer fh.Close()
	assert.NoError(t, png.Encode(fh, m.AsImage()))
}

func testMakeColumns(cols, height, col_width int,
	cb func(x, y int) float64) (out []Column) {

	for i := 0; i < cols; i++ {
		var data []float64
		for j := 0; j < height; j++ {
			data = append(data, cb(i, j))
		}
		out = append(out, Column{
			X:    i * col_width,
			W:    col_width,
			Data: data,
		})
	}

	return out
}
