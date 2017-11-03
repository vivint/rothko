// Copyright (C) 2017. See AUTHORS.

package draw

// TODO(jeff): make some []color.RGBA for people

var dumb []Color

func init() {
	for i := 0; i < 256; i++ {
		dumb = append(dumb, Color{
			R: uint8(i), G: uint8(i), B: uint8(i),
		})
	}
}
