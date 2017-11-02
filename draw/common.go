// Copyright (C) 2017. See AUTHORS.

package draw

import (
	"image"
	"image/color"
	"math"
)

// fastSet elides most bounds checks, assumes the rectangle for the image is
// at 0, 0 and ignores the alpha component.
func fastSet(m *image.RGBA, x, y int, c color.RGBA) {
	i := y*m.Stride + x*4
	_ = m.Pix[i+3]
	m.Pix[i+0] = c.R
	m.Pix[i+1] = c.G
	m.Pix[i+2] = c.B
	m.Pix[i+3] = 255
}

func fastFloor(f float64) int {
	y := math.Float64bits(f)
	e := 0x3ff + 63 - (y >> 52)
	m := 1<<63 | y<<11
	return int(m >> e)
}
