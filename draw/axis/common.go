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
		Rect:   image.Rect(0, 0, m.Width, m.Height),
	}
}
