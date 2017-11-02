// Copyright (C) 2017. See AUTHORS.

package draw

import "image/color"

// TODO(jeff): make some []color.RGBA for people

var dumb []color.RGBA

func init() {
	for i := 0; i < 256; i++ {
		dumb = append(dumb, color.RGBA{
			R: uint8(i), G: uint8(i), B: uint8(i), A: 255,
		})
	}
}
