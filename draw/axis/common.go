// Copyright (C) 2017. See AUTHORS.

package axis

import (
	"image"

	"github.com/spacemonkeygo/rothko/draw"
)

func asImage(m *draw.RGB) *image.RGBA {
	return &image.RGBA{
		Pix:    m.Pix,
		Stride: m.Stride,
		// TODO(jeff): i highly suspect m.Width and m.Height is wrong here.
		// add a test around boundary conditions.
		Rect: image.Rect(-m.X, -m.Y, m.Width, m.Height),
	}
}
