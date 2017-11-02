// Copyright (C) 2017. See AUTHORS.

package draw

import "image/color"

type BlendMode func(p float64, left, right color.RGBA) color.RGBA

func BlendModeLeft(p float64, left, right color.RGBA) color.RGBA {
	return left
}

func BlendModeLinear(p float64, left, right color.RGBA) color.RGBA {
	round := func(x float64) uint8 {
		return uint8(int(x*2) / 2)
	}

	blend := func(p float64, left, right uint8) uint8 {
		return round(p*float64(left) + (1-p)*float64(right))
	}

	return color.RGBA{
		R: blend(p, left.R, right.R),
		G: blend(p, left.G, right.G),
		B: blend(p, left.B, right.B),
		A: 255,
	}
}
